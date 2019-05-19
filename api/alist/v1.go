package alist

func NewTypeV1() *Alist {
	aList := &Alist{}

	aList.Info.ListType = SimpleList
	data := make(AlistTypeV1, 0)
	aList.Data = data

	labels := make([]string, 0)
	aList.Info.Labels = labels

	return aList
}
