import os
import pandas as pd
import numpy as np
from datetime import datetime
from sqlalchemy import create_engine, text
from neo4j import GraphDatabase
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# ==========================================
# DATABASE CONNECTION FROM ENVIRONMENT VARIABLES
# ==========================================
# PostgreSQL Config
DB_CONNECTION = os.getenv("DB_CONNECTION")
DB_HOST = os.getenv("DB_HOST")
DB_PORT = os.getenv("DB_PORT")
DB_DATABASE = os.getenv("DB_DATABASE")
DB_USERNAME = os.getenv("DB_USERNAME")
DB_PASSWORD = os.getenv("DB_PASSWORD")

POSTGRES_URI = f"postgresql://{DB_USERNAME}:{DB_PASSWORD}@{DB_HOST}:{DB_PORT}/{DB_DATABASE}"

# Neo4j Config
NEO4J_URI = os.getenv("NEO4J_URI")
NEO4J_USER = os.getenv("NEO4J_USER")
NEO4J_PASSWORD = os.getenv("NEO4J_PASSWORD")


def run_etl(file_path):
    now_str = datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%S.%fZ')

    # --------------------------------------
    # 1. EXTRACT DATA FROM KAGGLE CSV
    # --------------------------------------
    print("Extracting data from Kaggle dataset...")
    df_raw = pd.read_csv(file_path, usecols=['case:concept:name', 'concept:name', 'org:resource', 'time:timestamp'])
    df = df_raw.head(50000).copy() 
    df = df.rename(columns={
        'case:concept:name': 'case_id',
        'concept:name': 'Activity',
        'org:resource': 'Resource',
        'time:timestamp': 'StartTime'
    })

    df['case_numeric_id'] = pd.factorize(df['case_id'])[0] + 100000
    df['job_global_id'] = range(1, len(df) + 1)

    # --------------------------------------
    # 2. TRANSFORM METADATA (DIVISIONS & TITLES)
    # --------------------------------------
    print("Transforming Division and Title data (Fixed Schema)...")
    divisions_list = ['Information Technology', 'Finance & Accounting', 'Credit Operations', 'Legal']
    
    df_divisions = pd.DataFrame({
        'id': range(1, len(divisions_list) + 1),
        'name': divisions_list,
        'code': [f"DIV-{d.split()[0].upper()}" for d in divisions_list],
        'created_at': now_str,
        'updated_at': now_str,
        'deleted_at': None
    })

    unique_resources = df['Resource'].dropna().unique()
    titles_data = []
    resource_to_title = {}
    
    for idx, res in enumerate(unique_resources):
        t_id = idx + 1
        resource_to_title[res] = t_id
        # PERBAIKAN: division_id dihapus total dari entitas Title sesuai skema target
        titles_data.append({
            'id': t_id,
            'code': f"TTL-{idx+1:03d}",
            'color': "#FFFFFF",
            'name': f"Officer {res}",
            'created_at': now_str,
            'updated_at': now_str,
            'deleted_at': None
        })
    df_titles = pd.DataFrame(titles_data)
    df['title_id'] = df['Resource'].map(resource_to_title).fillna(1).astype(int)

    # --------------------------------------
    # 3. TRANSFORM FLOWCHARTS
    # --------------------------------------
    df_flowcharts = pd.DataFrame([
        {'id': 1, 'type': 'process'},
        {'id': 2, 'type': 'decision'}
    ])

    # --------------------------------------
    # 4. SEPARATE SOP AND SPK CASES (80:20 Split)
    # --------------------------------------
    print("Splitting procedures into SOPs and SPKs...")
    unique_cases = df['case_numeric_id'].unique()
    np.random.seed(42)
    np.random.shuffle(unique_cases)
    
    split_idx = int(len(unique_cases) * 0.8)
    sop_ids = unique_cases[:split_idx]
    spk_ids = unique_cases[split_idx:]

    df_sops = pd.DataFrame({
        'id': sop_ids,
        'name': [f"Prosedur Evaluasi Kredit {i}" for i in sop_ids],
        'code': [f"SOP-CR-{i}" for i in sop_ids],
        'description': "Master SOP dipetakan dari log aktivitas BPI 2017",
        'created_at': now_str,
        'updated_at': now_str,
        'deleted_at': None,
        'parent_job_id': None
    })

    df_spks = pd.DataFrame({
        'id': spk_ids,
        'name': [f"Surat Perintah Kerja Finansial {i}" for i in spk_ids],
        'code': [f"SPK-FIN-{i}" for i in spk_ids],
        'description': "SPK operasional pembiayaan internal",
        'created_at': now_str,
        'updated_at': now_str,
        'deleted_at': None
    })

    # --------------------------------------
    # 5. TRANSFORM JOBS WITH LINKED-LIST INDEXING
    # --------------------------------------
    print("Calculating Step Order Indexes (Next/Prev)...")
    df = df.sort_values(['case_numeric_id', 'StartTime'])
    df['index'] = df.groupby('case_numeric_id').cumcount() + 1
    df['prev_index'] = df.groupby('case_numeric_id')['index'].shift(1).astype('Int64')
    df['next_index'] = df.groupby('case_numeric_id')['index'].shift(-1).astype('Int64')

    df_sop_rows = df[df['case_numeric_id'].isin(sop_ids)].copy()
    df_spk_rows = df[df['case_numeric_id'].isin(spk_ids)].copy()

    # Logika Polimorfik SOP Job (Rujukan Antar Job)
    all_job_ids = df['job_global_id'].values
    def generate_sop_refs(row):
        rand = np.random.rand()
        if rand < 0.70:
            return 'instruction', None
        elif rand < 0.90:
            return 'spk', int(np.random.choice(all_job_ids))
        else:
            return 'sop', int(np.random.choice(all_job_ids))

    if not df_sop_rows.empty:
        refs = df_sop_rows.apply(generate_sop_refs, axis=1)
        df_sop_rows['type'] = [r[0] for r in refs]
        df_sop_rows['reference_id'] = [r[1] for r in refs]
        df_sop_rows['reference_id'] = df_sop_rows['reference_id'].astype('Int64')
    else:
        df_sop_rows['type'] = None
        df_sop_rows['reference_id'] = None

    df_sop_jobs = pd.DataFrame({
        'id': df_sop_rows['job_global_id'],
        'name': df_sop_rows['Activity'],
        'alias': "SOP_STEP_ALIAS",
        'type': df_sop_rows['type'],
        'code': "JOB-SOP-" + df_sop_rows['job_global_id'].astype(str),
        'description': "Langkah sekuensial master prosedur",
        'title_id': df_sop_rows['title_id'],
        'sop_id': df_sop_rows['case_numeric_id'],
        'reference_id': df_sop_rows['reference_id'],
        'index': df_sop_rows['index'],
        'is_published': True,
        'is_hide': False,
        'flowchart_id': 1,
        'next_index': df_sop_rows['next_index'],
        'prev_index': df_sop_rows['prev_index'],
        'created_at': now_str,
        'updated_at': now_str
    })

    def generate_spk_sop_ref(row):
        return int(np.random.choice(sop_ids)) if np.random.rand() < 0.20 else None

    if not df_spk_rows.empty:
        df_spk_rows['sop_id_ref'] = df_spk_rows.apply(generate_spk_sop_ref, axis=1).astype('Int64')
    else:
        df_spk_rows['sop_id_ref'] = None

    df_spk_jobs = pd.DataFrame({
        'id': df_spk_rows['job_global_id'],
        'name': df_spk_rows['Activity'],
        'description': "Langkah perintah kerja eksekusi lapangan",
        'spk_id': df_spk_rows['case_numeric_id'],
        'sop_id': df_spk_rows['sop_id_ref'],
        'title_id': df_spk_rows['title_id'],
        'index': df_spk_rows['index'],
        'flowchart_id': 2,
        'next_index': df_spk_rows['next_index'],
        'prev_index': df_spk_rows['prev_index'],
        'created_at': now_str,
        'updated_at': now_str
    })

    # Pembersihan tipe data NaN -> None
    df_divisions = df_divisions.replace({np.nan: None})
    df_titles = df_titles.replace({np.nan: None})
    df_sops = df_sops.replace({np.nan: None})
    df_spks = df_spks.replace({np.nan: None})
    df_sop_jobs = df_sop_jobs.replace({np.nan: None})
    df_spk_jobs = df_spk_jobs.replace({np.nan: None})

    # --------------------------------------
    # 6. INGEST TO POSTGRESQL
    # --------------------------------------
    print(f"Loading data into PostgreSQL database: {DB_DATABASE}...")
    engine = create_engine(POSTGRES_URI)
    
    # Hapus data child tables dulu (yang punya foreign key), lalu parent tables
    with engine.connect() as conn:
        conn.execute(text("TRUNCATE TABLE spk_titles, sop_divisions, user_has_divisions, spk_jobs, sop_jobs, spks, sops, titles, divisions RESTART IDENTITY CASCADE"))
        conn.commit()
    
    df_divisions.to_sql('divisions', engine, if_exists='append', index=False)
    df_titles[['id', 'code', 'color', 'name', 'created_at', 'updated_at', 'deleted_at']].to_sql('titles', engine, if_exists='append', index=False)
    
    # Flowcharts - Insert only if not exists
    try:
        with engine.connect() as conn:
            result = conn.execute(text("SELECT COUNT(*) FROM flowcharts WHERE id IN (1, 2)"))
            count = result.scalar()
            if count == 0:
                df_flowcharts.to_sql('flowcharts', engine, if_exists='append', index=False)
                print("Flowcharts data inserted.")
            else:
                print("Flowcharts data already exists, skipping.")
    except Exception as e:
        print(f"Flowcharts table not found, creating and inserting: {e}")
        df_flowcharts.to_sql('flowcharts', engine, if_exists='append', index=False)
    
    df_sops.to_sql('sops', engine, if_exists='append', index=False)
    df_spks.to_sql('spks', engine, if_exists='append', index=False)
    df_sop_jobs.to_sql('sop_jobs', engine, if_exists='append', index=False)
    df_spk_jobs.to_sql('spk_jobs', engine, if_exists='append', index=False)
    
    # --------------------------------------
    # PIVOT TABLES (Many-to-Many Relationships)
    # --------------------------------------
    print("Inserting pivot table data...")
    
    # sop_divisions: mapping each SOP to random divisions
    division_ids = df_divisions['id'].tolist()
    np.random.seed(42)
    sop_divisions = []
    for sop_id in sop_ids:
        num_divs = np.random.randint(1, 3)
        chosen_divs = np.random.choice(division_ids, size=num_divs, replace=False)
        for div_id in chosen_divs:
            sop_divisions.append({'sop_id': int(sop_id), 'division_id': int(div_id)})
    df_sop_divisions = pd.DataFrame(sop_divisions)
    df_sop_divisions.to_sql('sop_divisions', engine, if_exists='append', index=False)
    
    # spk_titles: mapping each SPK to random titles
    title_ids = df_titles['id'].tolist()
    spk_titles = []
    for spk_id in spk_ids:
        num_titles = np.random.randint(1, 3)
        chosen_titles = np.random.choice(title_ids, size=num_titles, replace=False)
        for title_id in chosen_titles:
            spk_titles.append({'spk_id': int(spk_id), 'title_id': int(title_id)})
    df_spk_titles = pd.DataFrame(spk_titles)
    df_spk_titles.to_sql('spk_titles', engine, if_exists='append', index=False)
    
    print(f"  - sop_divisions: {len(df_sop_divisions)} rows")
    print(f"  - spk_titles: {len(df_spk_titles)} rows")
    print("PostgreSQL loading completed.")

    # --------------------------------------
    # 7. INGEST TO NEO4J (EFFICIENT UNWIND)
    # --------------------------------------
    print(f"Loading data into Neo4j Instance: {NEO4J_URI}...")
    driver = GraphDatabase.driver(NEO4J_URI, auth=(NEO4J_USER, NEO4J_PASSWORD))
    
    with driver.session() as session:
        session.run("MATCH (n) DETACH DELETE n")

        # Ingest Node Flowchart (only if not exists)
        result = session.run("MATCH (f:Flowchart) RETURN count(f) as count")
        count = result.single()['count']
        if count == 0:
            session.run("""
                UNWIND $rows AS row
                CREATE (:Flowchart {id: row.id, type: row.type})
            """, rows=df_flowcharts.to_dict(orient='records'))
            print("Neo4j Flowcharts inserted.")
        else:
            print("Neo4j Flowcharts already exist, skipping.")

        # Ingest Node Division
        session.run("""
            UNWIND $rows AS row
            CREATE (:Division {
                id: row.id, name: row.name, code: row.code, 
                created_at: row.created_at, updated_at: row.updated_at, deleted_at: row.deleted_at
            })
        """, rows=df_divisions.to_dict(orient='records'))

        # Ingest Node Title (Tanpa properti division_id)
        session.run("""
            UNWIND $rows AS row
            CREATE (:Title {
                id: row.id, name: row.name, code: row.code, color: row.color,
                created_at: row.created_at, updated_at: row.updated_at, deleted_at: row.deleted_at
            })
        """, rows=df_titles.to_dict(orient='records'))

        # Ingest Node SOP
        session.run("""
            UNWIND $rows AS row
            CREATE (:SOP {
                id: row.id, name: row.name, code: row.code, description: row.description,
                parent_job_id: row.parent_job_id, created_at: row.created_at, 
                updated_at: row.updated_at, deleted_at: row.deleted_at
            })
        """, rows=df_sops.to_dict(orient='records'))

        # Ingest Node SPK
        session.run("""
            UNWIND $rows AS row
            CREATE (:SPK {
                id: row.id, name: row.name, code: row.code, description: row.description,
                created_at: row.created_at, updated_at: row.updated_at, deleted_at: row.deleted_at
            })
        """, rows=df_spks.to_dict(orient='records'))

        # Ingest Node Job
        jobs_combined = []
        for _, r in df_sop_jobs.iterrows():
            jobs_combined.append({
                'id': r['id'], 'name': r['name'], 'alias': r['alias'], 'type': r['type'],
                'code': r['code'], 'description': r['description'], 'title_id': r['title_id'],
                'sop_id': r['sop_id'], 'spk_id': None, 'reference_id': r['reference_id'],
                'index': r['index'], 'flowchart_id': r['flowchart_id'], 'next_index': r['next_index'],
                'prev_index': r['prev_index'], 'is_published': r['is_published'], 'is_hide': r['is_hide'],
                'created_at': r['created_at'], 'updated_at': r['updated_at'], 'deleted_at': None,
                'title_name': f"Officer Type {r['title_id']}", 'reference_name': "REF_JOB"
            })
        for _, r in df_spk_jobs.iterrows():
            jobs_combined.append({
                'id': r['id'], 'name': r['name'], 'alias': None, 'type': 'instruction',
                'description': r['description'], 'title_id': r['title_id'],
                'sop_id': r['sop_id'], 'spk_id': r['spk_id'], 'reference_id': None,
                'index': r['index'], 'flowchart_id': r['flowchart_id'], 'next_index': r['next_index'],
                'prev_index': r['prev_index'], 'is_published': True, 'is_hide': False,
                'created_at': r['created_at'], 'updated_at': r['updated_at'], 'deleted_at': None,
                'title_name': f"Officer Type {r['title_id']}", 'reference_name': None
            })
        
        session.run("""
            UNWIND $rows AS row
            CREATE (:Job {
                id: row.id, name: row.name, alias: row.alias, type: row.type, code: row.code,
                description: row.description, title_id: row.title_id, title_name: row.title_name,
                sop_id: row.sop_id, spk_id: row.spk_id, reference_id: row.reference_id,
                reference_name: row.reference_name, index: row.index, flowchart_id: row.flowchart_id,
                next_index: row.next_index, prev_index: row.prev_index, is_published: row.is_published,
                is_hide: row.is_hide, created_at: row.created_at, updated_at: row.updated_at, deleted_at: row.deleted_at
            })
        """, rows=jobs_combined)

        # --------------------------------------
        # 8. BUILD GRAPH RELATIONSHIPS
        # --------------------------------------
        print("Connecting Graph Relationships...")
        
        session.run("MATCH (s:SOP), (j:Job) WHERE j.sop_id = s.id AND j.spk_id IS NULL CREATE (s)-[:HAS_JOB]->(j)")
        session.run("MATCH (spk:SPK), (j:Job) WHERE j.spk_id = spk.id CREATE (spk)-[:HAS_JOB]->(j)")
        
        sop_divisions_payload = [{'sop_id': int(sid), 'division_id': int((sid % 4) + 1)} for sid in sop_ids]
        session.run("""
            UNWIND $pairs AS pair
            MATCH (s:SOP {id: pair.sop_id}), (d:Division {id: pair.division_id})
            CREATE (s)-[:BELONGS_TO]->(d)
        """, pairs=sop_divisions_payload)

        session.run("MATCH (j:Job), (t:Title) WHERE j.title_id = t.id CREATE (j)-[:HAS_TITLE]->(t)")
        session.run("MATCH (j1:Job), (j2:Job) WHERE j1.reference_id = j2.id CREATE (j1)-[:REFERENCES]->(j2)")

    driver.close()
    print("Neo4j loading completed.")
    print("=== ETL PIPELINE SUCCESSFUL ===")
    
if __name__ == "__main__":
    run_etl('pkg/dataset/bpi_2017_cleaned.csv')