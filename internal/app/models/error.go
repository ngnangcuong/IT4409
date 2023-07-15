package models

import "errors"

var ErrNoBlog = errors.New("blog is not exist")
var ErrNoComment = errors.New("comment is not exist")
var ErrNoPermission = errors.New("user does not have permission")
var ErrInvalidParameter = errors.New("invalid paramter")
var ErrInternalServerError = errors.New("something went wrong in server, you can try again")
