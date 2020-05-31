/*
  author='du'
  date='2020/5/31 10:09'
*/
package common

import "errors"

var (
	ERR_LOCK_ALREADY_OCCUPIED = errors.New("锁已经被占用了")
)
