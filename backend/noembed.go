//go:build !embed

package main

import "net/http"

// FrontendFS is nil when building without the embed tag.
// In this mode, the server serves only API endpoints.
var FrontendFS http.FileSystem
