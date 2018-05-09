package server

import (
	"fmt"
	"io/ioutil"
	"strings"

	"core.db"
	"core.globals"
	"core.vm"
	"github.com/labstack/echo"
)

// hello .
func hello(c echo.Context) error {
	return c.JSON(200, map[string]interface{}{
		"success": true,
		"data":    "Let's AggreX!",
	})
}

// quickExec .
func quickExec(c echo.Context) error {
	vm := vm.New(vm.VM{
		AllowedHosts: strings.Split(*globals.FlagAllowedHosts, ","),
		MaxExecTime:  *globals.FlagMaxExecTime,
		Request:      c.Request(),
	})

	code, _ := ioutil.ReadAll(c.Request().Body)
	if len(code) < 1 {
		return c.JSON(422, map[string]interface{}{
			"success": false,
			"error":   "Nothing to run!",
		})
	}

	result, err := vm.Exec(string(code))
	if err != nil {
		return c.JSON(500, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	if result == nil && vm.LastError != nil {
		return c.JSON(408, map[string]interface{}{
			"success": false,
			"error":   vm.LastError,
		})
	}
	return c.JSON(200, result)
}

// procedureSave .
func procedureSave(c echo.Context) error {
	code, _ := ioutil.ReadAll(c.Request().Body)

	procedure := db.Procedure{
		Key:  c.Param("key"),
		Code: string(code),
		Tags: strings.Split(c.QueryParam("tags"), ","),
	}

	globals.DBHandler.Put(&procedure)

	return c.JSON(200, map[string]interface{}{
		"success": true,
		"data":    procedure,
	})
}

// procedureSearch .
func procedureSearch(c echo.Context) error {
	var input struct {
		Query  string   `json:"query"`
		Sort   []string `json:"sort"`
		Offset int      `json:"offset"`
		Limit  int      `json:"limit"`
	}
	if err := c.Bind(&input); err != nil {
		return c.JSON(422, map[string]interface{}{
			"success": false,
			"errors": []string{
				err.Error(),
			},
		})
	}
	if input.Query == "" {
		input.Query = "+key: *"
	}
	result, err := globals.DBHandler.Find(input.Query, input.Sort, input.Offset, input.Limit)
	errs := []string{}
	if err != nil {
		errs = append(errs, err.Error())
	}
	return c.JSON(200, map[string]interface{}{
		"success": err == nil,
		"data":    result,
		"errors":  errs,
	})
}

// procedureExec .
func procedureExec(c echo.Context) error {
	vm := vm.New(vm.VM{
		AllowedHosts: strings.Split(*globals.FlagAllowedHosts, ","),
		MaxExecTime:  *globals.FlagMaxExecTime,
		Request:      c.Request(),
	})
	procedures, err := globals.DBHandler.Find(fmt.Sprintf("+key: '%s'", c.Param("key")), []string{}, 0, 1)
	if err != nil || procedures.Totals < 1 {
		return c.JSON(404, map[string]interface{}{
			"success": false,
			"error":   "procedure not found",
		})
	}
	result, err := vm.Exec(procedures.Hits[0]["code"].(string))
	if err != nil {
		return c.JSON(500, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	if result == nil && vm.LastError != nil {
		return c.JSON(408, map[string]interface{}{
			"success": false,
			"error":   vm.LastError,
		})
	}
	return c.JSON(200, result)
}

// procedureDelete .
func procedureDelete(c echo.Context) error {
	globals.DBHandler.Delete(c.Param("key"))
	return c.JSON(200, map[string]interface{}{
		"success": true,
	})
}

// globalsSet .
func globalsSet(c echo.Context) error {
	var vars map[string]interface{}
	if err := c.Bind(&vars); err != nil {
		return c.JSON(422, map[string]interface{}{
			"success": false,
			"errors": []string{
				err.Error(),
			},
		})
	}
	globals.DBHandler.GlobalsSet(vars)
	return c.JSON(200, map[string]interface{}{
		"success": true,
		"data":    globals.DBHandler.GlobalsGet(),
	})
}

// globalsGet .
func globalsGet(c echo.Context) error {
	return c.JSON(200, map[string]interface{}{
		"success": true,
		"data":    globals.DBHandler.GlobalsGet(),
	})
}
