package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	IPAddressKey contextKey = "ip_address"
	UserAgentKey contextKey = "user_agent"
)

const TableNameActivityLog = "activity_logs"

var skipTables = []string{
	"case_report_categories",
	"case_report_bug_features",
	"case_report_technicians",
	"case_report_assigned_users",
	"case_reports",
	"squad_members",
	"user_has_roles",
}

func shouldSkipTable(tableName string) bool {
	for _, skipTable := range skipTables {
		if tableName == skipTable {
			return true
		}
	}
	return false
}

type AuditLoggerPlugin struct{}

func (p *AuditLoggerPlugin) Name() string {
	return "AuditLoggerPlugin"
}

func (p *AuditLoggerPlugin) Initialize(db *gorm.DB) error {
	if err := db.Callback().Create().After("gorm:create").Register("audit:after_create", p.afterCreate); err != nil {
		return err
	}

	if err := db.Callback().Update().Before("gorm:update").Register("audit:before_update", p.beforeUpdate); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").Register("audit:after_update", p.afterUpdate); err != nil {
		return err
	}

	if err := db.Callback().Delete().Before("gorm:delete").Register("audit:before_delete", p.beforeDelete); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gorm:delete").Register("audit:after_delete", p.afterDelete); err != nil {
		return err
	}

	return nil
}

func (p *AuditLoggerPlugin) afterCreate(db *gorm.DB) {
	if db.Error != nil || db.Statement.Schema == nil {
		return
	}

	if db.Statement.Schema.Table == TableNameActivityLog {
		return
	}

	if shouldSkipTable(db.Statement.Schema.Table) {
		return
	}

	value := db.Statement.ReflectValue
	if !value.IsValid() {
		return
	}

	newData, err := serializeData(db.Statement.Dest)
	if err != nil {
		logError("Failed to serialize new data for audit log: %v", err)
		return
	}

	ctx := db.Statement.Context

	if ctx == nil {
		return
	}

	userID := getUserIDFromContext(ctx)
	if userID == nil {
		return
	}

	ipAddress := getIPAddressFromContext(ctx)
	userAgent := getUserAgentFromContext(ctx)

	recordID := extractRecordID(db.Statement.Dest)

	if recordID == nil {
		return
	}

	auditLog := map[string]interface{}{
		"user_id":      userID,
		"table_ref":    db.Statement.Schema.Table,
		"table_ref_id": recordID,
		"action":       "CREATE",
		"message":      fmt.Sprintf("membuat data %s baru", db.Statement.Schema.Table),
		"new_data":     newData,
		"user_agent":   userAgent,
		"origins":      ipAddress,
		"created_at":   time.Now(),
	}

	if err := db.Session(&gorm.Session{NewDB: true}).Table(TableNameActivityLog).Create(auditLog).Error; err != nil {
		logError("Failed to create audit log: %v", err)
	}
}

func (p *AuditLoggerPlugin) beforeUpdate(db *gorm.DB) {
	if db.Error != nil || db.Statement.Schema == nil {
		return
	}

	if db.Statement.Schema.Table == TableNameActivityLog {
		return
	}

	if shouldSkipTable(db.Statement.Schema.Table) {
		return
	}

	if db.Statement.Context == nil {
		return
	}
	userID := getUserIDFromContext(db.Statement.Context)
	if userID == nil {
		return
	}

	var recordID *int64

	if db.Statement.Model != nil {
		recordID = extractRecordID(db.Statement.Model)
	}

	if recordID == nil && db.Statement.Dest != nil {
		recordID = extractRecordID(db.Statement.Dest)
	}

	if recordID == nil {
		return
	}

	modelType := reflect.New(db.Statement.Schema.ModelType).Interface()
	if err := db.Session(&gorm.Session{NewDB: true}).Where("id = ?", *recordID).First(modelType).Error; err == nil {
		db.Statement.Context = context.WithValue(db.Statement.Context, contextKey("old_record"), modelType)
	}
}

func (p *AuditLoggerPlugin) afterUpdate(db *gorm.DB) {
	if db.Error != nil || db.Statement.Schema == nil {
		return
	}

	if db.Statement.Schema.Table == TableNameActivityLog {
		return
	}

	if shouldSkipTable(db.Statement.Schema.Table) {
		return
	}

	if db.Statement.Context == nil {
		return
	}
	userID := getUserIDFromContext(db.Statement.Context)
	if userID == nil {
		return
	}

	oldRecord := db.Statement.Context.Value(contextKey("old_record"))
	if oldRecord == nil {
		return
	}

	var recordID *int64
	if db.Statement.Model != nil {
		recordID = extractRecordID(db.Statement.Model)
	}
	if recordID == nil && db.Statement.Dest != nil {
		recordID = extractRecordID(db.Statement.Dest)
	}
	if recordID == nil {
		return
	}

	newRecord := reflect.New(db.Statement.Schema.ModelType).Interface()
	if err := db.Session(&gorm.Session{NewDB: true}).Where("id = ?", *recordID).First(newRecord).Error; err != nil {
		logError("Failed to fetch updated record for audit log: %v", err)
		return
	}

	oldData, err := serializeData(oldRecord)
	if err != nil {
		logError("Failed to serialize old data for audit log: %v", err)
		return
	}

	newData, err := serializeData(newRecord)
	if err != nil {
		logError("Failed to serialize new data for audit log: %v", err)
		return
	}

	changes := calculateChanges(oldData, newData)

	if changes == nil || string(*changes) == "{}" {
		return
	}

	if isOnlyTimestampChanges(changes) {
		return
	}

	ctx := db.Statement.Context
	ipAddress := getIPAddressFromContext(ctx)
	userAgent := getUserAgentFromContext(ctx)

	auditLog := map[string]interface{}{
		"user_id":      userID,
		"table_ref":    db.Statement.Schema.Table,
		"table_ref_id": recordID,
		"action":       "UPDATE",
		"message":      fmt.Sprintf("mengupdate data %s", db.Statement.Schema.Table),
		"old_data":     oldData,
		"new_data":     newData,
		"changes":      changes,
		"user_agent":   userAgent,
		"origins":      ipAddress,
		"created_at":   time.Now(),
	}

	if err := db.Session(&gorm.Session{NewDB: true}).Table(TableNameActivityLog).Create(auditLog).Error; err != nil {
		logError("Failed to create audit log: %v", err)
	}
}

func (p *AuditLoggerPlugin) beforeDelete(db *gorm.DB) {
	if db.Error != nil || db.Statement.Schema == nil {
		return
	}

	if db.Statement.Schema.Table == TableNameActivityLog {
		return
	}

	if shouldSkipTable(db.Statement.Schema.Table) {
		return
	}

	if db.Statement.Context == nil {
		return
	}
	userID := getUserIDFromContext(db.Statement.Context)
	if userID == nil {
		return
	}

	if db.Statement.Dest != nil {
		recordID := extractRecordID(db.Statement.Dest)
		if recordID == nil {
			return
		}

		oldRecord := reflect.New(reflect.TypeOf(db.Statement.Dest).Elem()).Interface()
		if err := db.Session(&gorm.Session{NewDB: true}).Where("id = ?", *recordID).First(oldRecord).Error; err == nil {
			db.Statement.Context = context.WithValue(db.Statement.Context, contextKey("old_record"), oldRecord)
		}
	}
}

func (p *AuditLoggerPlugin) afterDelete(db *gorm.DB) {
	if db.Error != nil || db.Statement.Schema == nil {
		return
	}

	if db.Statement.Schema.Table == TableNameActivityLog {
		return
	}

	if shouldSkipTable(db.Statement.Schema.Table) {
		return
	}

	if db.Statement.Context == nil {
		return
	}
	userID := getUserIDFromContext(db.Statement.Context)
	if userID == nil {
		return
	}

	oldRecord := db.Statement.Context.Value(contextKey("old_record"))
	if oldRecord == nil {
		return
	}

	oldData, err := serializeData(oldRecord)
	if err != nil {
		logError("Failed to serialize old data for audit log: %v", err)
		return
	}

	ctx := db.Statement.Context
	ipAddress := getIPAddressFromContext(ctx)
	userAgent := getUserAgentFromContext(ctx)

	// Get record ID
	recordID := extractRecordID(oldRecord)

	// Skip if record ID is null
	if recordID == nil {
		return
	}

	// Create audit log entry
	auditLog := map[string]interface{}{
		"user_id":      userID,
		"table_ref":    db.Statement.Schema.Table,
		"table_ref_id": recordID,
		"action":       "DELETE",
		"message":      fmt.Sprintf("menghapus data dari %s", db.Statement.Schema.Table),
		"old_data":     oldData,
		"user_agent":   userAgent,
		"origins":      ipAddress,
		"created_at":   time.Now(),
	}

	if err := db.Session(&gorm.Session{NewDB: true}).Table(TableNameActivityLog).Create(auditLog).Error; err != nil {
		logError("Failed to create audit log: %v", err)
	}
}

// Helper functions
func serializeData(data interface{}) (*json.RawMessage, error) {
	if data == nil {
		return nil, nil
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	rawMsg := json.RawMessage(jsonBytes)
	return &rawMsg, nil
}

func extractRecordID(data interface{}) *int64 {
	if data == nil {
		return nil
	}

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	idField := val.FieldByName("ID")
	if !idField.IsValid() {
		return nil
	}

	switch idField.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		id := idField.Int()
		return &id
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		id := int64(idField.Uint())
		return &id
	}

	return nil
}

func calculateChanges(oldData, newData *json.RawMessage) *json.RawMessage {
	if oldData == nil || newData == nil {
		return nil
	}

	var oldMap, newMap map[string]interface{}
	if err := json.Unmarshal(*oldData, &oldMap); err != nil {
		return nil
	}
	if err := json.Unmarshal(*newData, &newMap); err != nil {
		return nil
	}

	changes := make(map[string]map[string]interface{})

	for key, newValue := range newMap {
		oldValue, exists := oldMap[key]
		if !exists || !reflect.DeepEqual(oldValue, newValue) {
			changes[key] = map[string]interface{}{
				"old": oldValue,
				"new": newValue,
			}
		}
	}

	if len(changes) == 0 {
		return nil
	}

	jsonBytes, err := json.Marshal(changes)
	if err != nil {
		return nil
	}

	rawMsg := json.RawMessage(jsonBytes)
	return &rawMsg
}

func isOnlyTimestampChanges(changes *json.RawMessage) bool {
	if changes == nil {
		return false
	}

	var changesMap map[string]interface{}
	if err := json.Unmarshal(*changes, &changesMap); err != nil {
		return false
	}

	timestampFields := map[string]bool{
		"updated_at": true,
		"created_at": true,
		"deleted_at": true,
	}

	for field := range changesMap {
		if !timestampFields[field] {
			return false
		}
	}

	return true
}

func getUserIDFromContext(ctx context.Context) *int64 {
	if userID, ok := ctx.Value(UserIDKey).(int64); ok {
		return &userID
	}
	return nil
}

func getIPAddressFromContext(ctx context.Context) *string {
	if ip, ok := ctx.Value(IPAddressKey).(string); ok {
		return &ip
	}
	return nil
}

func getUserAgentFromContext(ctx context.Context) *string {
	if ua, ok := ctx.Value(UserAgentKey).(string); ok {
		return &ua
	}
	return nil
}

func logError(format string, args ...interface{}) {
	fmt.Printf("[AUDIT ERROR] "+format+"\n", args...)
}
