package alist

func NewTypeV2() *Alist {
	aList := &Alist{}

	aList.Info.ListType = FromToList
	data := make(AlistTypeV2, 0)
	aList.Data = data

	labels := make([]string, 0)
	aList.Info.Labels = labels

	return aList
}
