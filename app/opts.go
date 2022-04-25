package app

import (
	"strings"
)

// CommonOpts is the options that provided into handlers
type CommonOpts struct {
	AppURL     string
	BoltPath   string
	StaticPath string
	TmlPath    string
	TplExt     string
}

// SetCommon apply the options
func (c *CommonOpts) SetCommon(commonOpts CommonOpts) {
	c.AppURL = strings.TrimSuffix(commonOpts.AppURL, "/")
	c.BoltPath = strings.TrimSuffix(commonOpts.BoltPath, "/")
	c.StaticPath = strings.TrimSuffix(commonOpts.StaticPath, "/")
	c.TmlPath = strings.TrimSuffix(commonOpts.TmlPath, "/")
	c.TplExt = commonOpts.TplExt
}
