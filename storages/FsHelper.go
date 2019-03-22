package storages

var _fs_ Fs

func SetFs(newFs Fs) {
	if newFs != nil {
		_fs_ = newFs
	}
}

func GetFs() Fs {
	if _fs_ == nil {
		_fs_ = NewOsFs()
	}
	return _fs_
}
