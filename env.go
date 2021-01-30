package main

import (
	"os"
	"strconv"
)

//PageAccessToken -
var PageAccessToken = os.Getenv(`PageAccessToken`)

//VerifyToken -
var VerifyToken = os.Getenv(`VerifyToken`)

//BasicAuth -
var BasicAuth = os.Getenv(`BasicAuth`)

//GistID -
var GistID = os.Getenv(`GistID`)

//MaxAgeDiff -
var MaxAgeDiff, _ = strconv.Atoi(os.Getenv(`MaxAgeDiff`))

//MaxAttempt -
var MaxAttempt, _ = strconv.Atoi(os.Getenv(`MaxAttempt`))

//Limit -
var Limit, _ = strconv.Atoi(os.Getenv(`Limit`))
