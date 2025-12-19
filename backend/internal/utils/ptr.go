package utils

// Int32PtrToIntPtr converts a *int32 to *int
func Int32PtrToIntPtr(i *int32) *int {
	if i == nil {
		return nil
	}
	v := int(*i)
	return &v
}
