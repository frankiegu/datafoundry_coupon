package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Coupon struct {
	Id         int
	Serial     string    `json:"serial"`
	Code       string    `json:"code"`
	Kind       string    `json:"kind,omitempty"`
	Expiration time.Time `json:"expiration,omitempty"`
	Region     string    `json:"region,omitempty"`
	Amount     float32   `json:"amount,omitempty"`
}

type PlanRegion struct {
	Region          string `json:"region"`
	Region_describe string `json:"region_describe"`
	Identification  string `json:"identification"`
}

type Result struct {
	Id              int       `json:"id,omitempty"`
	Plan_id         string    `json:"plan_id,omitempty"`
	Plan_name       string    `json:"plan_name,omitempty"`
	Plan_type       string    `json:"plan_type,omitempty"`
	Plan_level      int       `json:"plan_level,omitempty"`
	Specification1  string    `json:"specification1,omitempty"`
	Specification2  string    `json:"specification2,omitempty"`
	Price           float32   `json:"price,omitempty"`
	Cycle           string    `json:"cycle,omitempty"`
	Region          string    `json:"region,omitempty"`
	Region_describe string    `json:"region_describe,omitempty"`
	Create_time     time.Time `json:"creation_time,omitempty"`
	Status          string    `json:"status,omitempty"`
}

type UseInfo struct {
	Serial    string    `json:"serial"`
	Code      string    `json:"code"`
	Username  string    `json:"username"`
	Namespace string    `json:"namespace"`
	Use_time  time.Time `json:"recharge_time"`
}

type createResult struct {
	Serial string `json:"serial"`
	Code   string `json:"code"`
}

func CreateCoupon(db *sql.DB, couponInfo *Coupon) (createResult, error) {
	logger.Info("Begin create a Coupon model.")

	sqlstr := fmt.Sprintf(`insert into DF_COUPON (
				SERIAL, CODE, KIND, EXPIRATION, REGION, AMOUNT, STATUS
				) values (?, ?, ?, ?, ?, ?, ?)`,
	)

	couponInfo.Serial = strings.ToLower(couponInfo.Serial)
	couponInfo.Code = strings.ToLower(couponInfo.Code)
	_, err := db.Exec(sqlstr,
		couponInfo.Serial, couponInfo.Code, couponInfo.Kind, couponInfo.Expiration,
		couponInfo.Region, couponInfo.Amount, "available",
	)

	result := createResult{Serial: couponInfo.Serial, Code: couponInfo.Code}

	logger.Info("End create a plan model.")
	return result, err
}

func DeletePlan(db *sql.DB, planId string) error {
	logger.Info("Model begin delete a plan.")

	err := modifyPlanStatusToN(db, planId)
	if err != nil {
		return err
	}

	logger.Info("Model begin delete a plan.")
	return err
}

//func ModifyPlan(db *sql.DB, planInfo *Coupon) error {
//	logger.Info("Model begin modify a plan.")
//	defer logger.Info("Model begin modify a plan.")
//
//	plan, err := RetrievePlanByID(db, planInfo.Plan_id)
//	if err != nil {
//		return err
//	} else if plan == nil {
//		return errors.New("Without this plan.")
//	}
//	logger.Debug("Retrieve plan: %v", plan)
//
//	err = modifyPlanStatusToN(db, plan.Plan_id)
//	if err != nil {
//		return err
//	}
//
//	_, err = CreatePlan(db, planInfo)
//	if err != nil {
//		return err
//	}
//
//	return err
//}

func RetrievePlanByID(db *sql.DB, planID string) (*Result, error) {
	logger.Info("Model begin get a plan by id.")

	logger.Info("Model end get a plan by id.")
	return getSinglePlan(db, fmt.Sprintf("PLAN_ID = '%s' and STATUS = 'A'", planID))
}

func getSinglePlan(db *sql.DB, sqlWhere string) (*Result, error) {
	apps, err := queryPlans(db, sqlWhere, 1, 0)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if len(apps) == 0 {
		return nil, nil
	}

	return apps[0], nil
}

func queryPlans(db *sql.DB, sqlWhere string, limit int, offset int64, sqlParams ...interface{}) ([]*Result, error) {
	offset_str := ""
	if offset > 0 {
		offset_str = fmt.Sprintf("offset %d", offset)
	}

	sqlWhereAll := ""
	if sqlWhere != "" {
		sqlWhereAll = fmt.Sprintf("WHERE P.REGION_ID=R.ID AND %s", sqlWhere)
	}

	sql_str := fmt.Sprintf(`select
					P.ID, P.PLAN_ID, P.PLAN_NAME,
					P.PLAN_TYPE, P.PLAN_LEVEL,
					P.SPECIFICATION1,
					P.SPECIFICATION2,
					P.PRICE, P.CYCLE,
					R.REGION, R.REGION_DESCRIBE,
					P.CREATE_TIME, P.STATUS
					from DF_PLAN P, DF_PLAN_REGION R
					%s
					limit %d
					%s
					`,
		sqlWhereAll,
		limit,
		offset_str)
	rows, err := db.Query(sql_str, sqlParams...)

	logger.Info(">>> %v", sql_str)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plans := make([]*Result, 0, 100)
	for rows.Next() {
		plan := &Result{}
		err := rows.Scan(
			&plan.Id, &plan.Plan_id, &plan.Plan_name, &plan.Plan_type, &plan.Plan_level, &plan.Specification1, &plan.Specification2,
			&plan.Price, &plan.Cycle, &plan.Region, &plan.Region_describe, &plan.Create_time, &plan.Status,
		)
		if err != nil {
			return nil, err
		}
		//validateApp(s) // already done in scanAppWithRows
		plans = append(plans, plan)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}

func modifyPlanStatusToN(db *sql.DB, planId string) error {
	sqlstr := fmt.Sprintf(`update DF_PLAN set status = "N" where PLAN_ID = '%s' and STATUS = 'A'`, planId)

	_, err := db.Exec(sqlstr)
	if err != nil {
		return err
	}

	return err
}

func QueryPlans(db *sql.DB, region, ptype, orderBy string, sortOrder bool, offset int64, limit int) (int64, []*Result, error) {
	logger.Info("Model begin get plan list.")
	defer logger.Info("Model end get plan list.")

	sqlParams := make([]interface{}, 0, 4)

	// ...

	sqlWhere := "STATUS = 'A'"
	region = strings.ToLower(region)
	if region != "" {

		regionId, err := getRegionId(db, region)
		if err != nil {
			return 0, nil, err
		}

		if sqlWhere == "" {
			sqlWhere = "REGION_ID = ?"
		} else {
			sqlWhere = sqlWhere + " and REGION_ID = ?"
		}
		sqlParams = append(sqlParams, regionId)
	}

	ptype = strings.ToLower(ptype)
	if ptype != "" {
		if sqlWhere == "" {
			sqlWhere = "PLAN_TYPE = ?"
		} else {
			sqlWhere = sqlWhere + " and PLAN_TYPE = ?"
		}
		sqlParams = append(sqlParams, ptype)
	}

	// ...

	switch strings.ToLower(orderBy) {
	default:
		orderBy = "CREATE_TIME"
		sortOrder = false
	case "createtime":
		orderBy = "CREATE_TIME"
	case "hotness":
		orderBy = "HOTNESS"
	}

	sqlSort := fmt.Sprintf("%s %s", orderBy, sortOrderText[sortOrder])

	// ...

	return getPlanList(db, offset, limit, sqlWhere, sqlSort, sqlParams...)
}

func getRegionId(db *sql.DB, region string) (int, error) {
	sql := `SELECT ID FROM DF_PLAN_REGION WHERE REGION=?`

	row := db.QueryRow(sql, region)

	var regionId int
	err := row.Scan(&regionId)
	if err != nil {
		return 0, err
	}

	return regionId, err
}

const (
	SortOrder_Asc  = "asc"
	SortOrder_Desc = "desc"
)

// true: asc
// false: desc
var sortOrderText = map[bool]string{true: "asc", false: "desc"}

func ValidateSortOrder(sortOrder string, defaultOrder bool) bool {
	switch strings.ToLower(sortOrder) {
	case SortOrder_Asc:
		return true
	case SortOrder_Desc:
		return false
	}

	return defaultOrder
}

func ValidateOrderBy(orderBy string) string {
	switch orderBy {
	case "createtime":
		return "CREATE_TIME"
	case "hotness":
		return "HOTNESS"
	}

	return ""
}

func getPlanList(db *sql.DB, offset int64, limit int, sqlWhere string, sqlSort string, sqlParams ...interface{}) (int64, []*Result, error) {
	//if strings.TrimSpace(sqlWhere) == "" {
	//	return 0, nil, errors.New("sqlWhere can't be blank")
	//}

	count, err := queryPlansCount(db, sqlWhere, sqlParams...)
	logger.Debug("count: %v", count)
	if err != nil {
		return 0, nil, err
	}
	if count == 0 {
		return 0, []*Result{}, nil
	}
	validateOffsetAndLimit(count, &offset, &limit)

	subs, err := queryPlans(db,
		fmt.Sprintf(`%s order by %s`, sqlWhere, sqlSort),
		limit, offset, sqlParams...)

	return count, subs, err
}

func queryPlansCount(db *sql.DB, sqlWhere string, sqlParams ...interface{}) (int64, error) {
	sqlWhere = strings.TrimSpace(sqlWhere)
	sql_where_all := ""
	if sqlWhere != "" {
		sql_where_all = fmt.Sprintf("where %s", sqlWhere)
	}

	count := int64(0)
	sql_str := fmt.Sprintf(`select COUNT(*) from DF_PLAN %s`, sql_where_all)
	logger.Debug(">>>\n"+
		"	%s", sql_str)
	logger.Debug("sqlParams: %v", sqlParams)
	err := db.QueryRow(sql_str, sqlParams...).Scan(&count)

	return count, err
}

func validateOffsetAndLimit(count int64, offset *int64, limit *int) {
	if *limit < 1 {
		*limit = 1
	}
	if *offset >= count {
		*offset = count - int64(*limit)
	}
	if *offset < 0 {
		*offset = 0
	}
	if *offset+int64(*limit) > count {
		*limit = int(count - *offset)
	}
}

func RetrievePlanRegion(db *sql.DB) ([]PlanRegion, error) {
	logger.Info("Model begin get plans region.")

	sql := "SELECT REGION, REGION_DESCRIBE, IDENTIFICATION FROM DF_PLAN_REGION"

	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}

	regions := make([]PlanRegion, 0)
	var region PlanRegion
	for rows.Next() {
		err = rows.Scan(&region.Region, &region.Region_describe, &region.Identification)
		if err != nil {
			return nil, err
		}

		regions = append(regions, region)
	}

	logger.Info("Model end get plan region.")

	return regions, err
}

type useResult struct {
	Amount float32
}

func UseCoupon(db *sql.DB, useInfo *UseInfo) (interface{}, error) {
	logger.Info("Begin use a coupon model.")

	useInfo.Serial = strings.ToLower(useInfo.Serial)
	useInfo.Code = strings.ToLower(useInfo.Code)

	sql := "SELECT AMOUNT, EXPIRATION, STATUS FROM DF_COUPON WHERE SERIAL=? AND CODE=?"
	row := db.QueryRow(sql, useInfo.Serial, useInfo.Code)
	logger.Info(">>>\n%v\n%v, %v", sql, useInfo.Serial, useInfo.Code)

	var amount float32
	var expiration time.Time
	var status string
	err := row.Scan(&amount, &expiration, &status)
	if err != nil {
		return nil, err
	}
	logger.Debug("expiration=%v, amount=%v, status=%v", expiration, amount, status)

	if status == "expired" {
		return nil, errors.New("The coupon has expired.")
	} else if status == "used" {
		return nil, errors.New("The coupon has used.")
	} else {
		return nil, errors.New("The coupon unavailable.")
	}

	useInfo.Use_time = useInfo.Use_time.UTC().Add(time.Hour * 8)
	logger.Debug("use time: %v", useInfo.Use_time)

	duration := expiration.Sub(useInfo.Use_time)
	logger.Debug("duration: %v", duration)

	if duration < 0 {
		sql = "UPDATE DF_COUPON SET STATUS='expired' WHERE SERIAL=? AND CODE=?"
		_, err := db.Exec(sql, useInfo.Serial, useInfo.Code)
		if err != nil {
			return nil, err
		}
		logger.Info(">>>\n%v\n%v, %v", sql, useInfo.Serial, useInfo.Code)
		return nil, errors.New("The coupon has expired.")
	}

	sql = "UPDATE DF_COUPON SET USE_TIME=?, USERNAME=?, NAMESPACE=?, STATUS=? WHERE SERIAL=? AND CODE=?"
	_, err = db.Exec(sql, useInfo.Use_time, useInfo.Username, useInfo.Namespace, "used", useInfo.Serial, useInfo.Code)
	if err != nil {
		return nil, err
	}
	logger.Info(">>>\n%v\n%v, %v, %v, %v, %v", sql,
		useInfo.Use_time, useInfo.Username, useInfo.Namespace, useInfo.Serial, useInfo.Code)

	useResult := useResult{Amount: amount}

	logger.Info("End use a coupon model.")
	return useResult, err
}