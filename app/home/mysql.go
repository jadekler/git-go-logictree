package home

import (
    "errors"
    "fmt"
    "github.com/jadekler/git-go-logictree/app/common"
)

func getMatchingUsers() ([]userSqlRow, error) {
    conditions := getConditions()

    sql := "SELECT name, age, num_pets FROM logictree.users WHERE "

    for key, condition := range conditions {
        if key != 0 {
            sql += " AND "
        }

        sql += condition.Field + " "

        switch condition.Operator {
        case "eq":
            sql += "="
        case "gt":
            sql += ">"
        case "lt":
            sql += "<"
        default:
            return nil, errors.New("Error: your conditions contain an operator that isn't legit - " + condition.Operator)
        }

        sql += " " + condition.Value
    }

    var name string
    var age, numPets int
    var userRowsReturned []userSqlRow

    rows, err := common.DB.Query(sql)
    common.CheckError(err, 2)
    defer rows.Close()

    for rows.Next() {
        rows.Scan(&name, &age, &numPets)
        userRowsReturned = append(userRowsReturned, userSqlRow{Name: name, Age: age, NumPets: numPets})
    }

    return userRowsReturned, nil
}

func getUserSqlRows() []userSqlRow {
    var name string
    var age, numPets int
    var userRowsReturned []userSqlRow

    // Get equality sql rows
    rows, err := common.DB.Query("SELECT name, age, num_pets FROM logictree.users")
    common.CheckError(err, 2)
    defer rows.Close()

    for rows.Next() {
        rows.Scan(&name, &age, &numPets)
        userRowsReturned = append(userRowsReturned, userSqlRow{Name: name, Age: age, NumPets: numPets})
    }

    return userRowsReturned
}

func (t *treeNode) toMysql() (equalityStr, logicStr string, err error) {
    t.attachLeftsAndRights()

    equalityStr, logicStr = t.toMysqlRecursively()

    if len(equalityStr) == 0 {
        return "", "", errors.New(fmt.Sprintf("Error: equality string was empty, which was unexpected. EqualityStr: %s :: LogicStr: %s", equalityStr, logicStr))
    }

    if len(logicStr) == 0 {
        return "", "", errors.New(fmt.Sprintf("Error: logic string was empty, which was unexpected. EqualityStr: %s :: LogicStr: %s", equalityStr, logicStr))
    }

    equalityStr = equalityStr[:(len(equalityStr)-1)]
    logicStr = logicStr[:(len(logicStr)-1)]

    return equalityStr, logicStr, nil
}

func (t *treeNode) toMysqlRecursively() (equalityStr, logicStr string) {
    var equalityTemp, logicTemp string
    for _, child := range t.Children {
        equalityTemp, logicTemp = child.toMysqlRecursively()
        equalityStr += equalityTemp
        logicStr += logicTemp
    }

    switch (t.Node.Type) {
    case "equality":
        equalityStr += fmt.Sprintf("('%s', '%s', '%s', 'equality', %d, %d),", t.Node.Field, t.Node.Operator, t.Node.Value, t.Left, t.Right)
    case "logic":
        logicStr += fmt.Sprintf("('%s', 'logic', %d, %d),", t.Node.Operator, t.Left, t.Right)
    }

    return equalityStr, logicStr
}

func updateDatabase(equalityStr, logicStr, usersStr string) {
    _, err := common.DB.Query("TRUNCATE TABLE logictree.conditions")
    common.CheckError(err, 2)

    _, err = common.DB.Query("TRUNCATE TABLE logictree.users")
    common.CheckError(err, 2)

    if equalityStr != "" {
        _, err = common.DB.Query("INSERT INTO logictree.conditions (field, operator, value, type, lt, rt) VALUES "+equalityStr)
        common.CheckError(err, 2)
    }

    if logicStr != "" {
        _, err = common.DB.Query("INSERT INTO logictree.conditions (operator, type, lt, rt) VALUES "+logicStr)
        common.CheckError(err, 2)
    }

    if usersStr != "" {
        _, err = common.DB.Query("INSERT INTO logictree.users (name, age, num_pets) VALUES "+usersStr)
        common.CheckError(err, 2)
    }
}

func getConditions() []conditionSqlRow {
    var Field, Operator, Value, Type string
    var Left, Right int
    var conditionRowsReturned []conditionSqlRow

    // Get equality sql rows
    rows, err := common.DB.Query("SELECT COALESCE(field, ''), operator, COALESCE(value, ''), type, lt, rt FROM logictree.conditions ORDER BY lt")
    common.CheckError(err, 2)
    defer rows.Close()

    for rows.Next() {
        rows.Scan(&Field, &Operator, &Value, &Type, &Left, &Right)
        conditionRowsReturned = append(conditionRowsReturned, conditionSqlRow{Field: Field, Operator: Operator, Value: Value, Type: Type, Left: Left, Right: Right})
    }

    return conditionRowsReturned
}