package rri

// QueryFieldList contains an ordered list of query fields.
type QueryFieldList []QueryField

// QueryField represents a single key-value pair of a query.
type QueryField struct {
	Name  QueryFieldName
	Value string
}

func newQueryFieldList() QueryFieldList {
	return make([]QueryField, 0)
}

// Size returns the number of values stored in this list.
func (l *QueryFieldList) Size() int {
	return len(*l)
}

// Add adds a sequence of values for the given field name.
func (l *QueryFieldList) Add(fieldName QueryFieldName, values ...string) {
	fieldName = fieldName.Normalize()
	for _, val := range values {
		*l = append(*l, QueryField{fieldName, val})
	}
}

// RemoveAll removes all values for a given field name
func (l *QueryFieldList) RemoveAll(fieldName QueryFieldName) {
	fieldName = fieldName.Normalize()
	newList := make([]QueryField, 0)
	for _, entry := range *l {
		if entry.Name != fieldName {
			newList = append(newList, entry)
		}
	}
	*l = newList
}

// Values returns all values defined for a field name.
func (l *QueryFieldList) Values(fieldName QueryFieldName) []string {
	fieldName = fieldName.Normalize()
	arr := make([]string, 0)
	for _, f := range *l {
		if f.Name == fieldName {
			arr = append(arr, f.Value)
		}
	}
	return arr
}

// FirstValue returns the first field value or an empty string for a field name.
func (l *QueryFieldList) FirstValue(fieldName QueryFieldName) string {
	fieldName = fieldName.Normalize()
	for _, f := range *l {
		if f.Name == fieldName {
			return f.Value
		}
	}
	return ""
}

// ResponseFieldList contains an ordered list of query fields.
type ResponseFieldList []ResponseField

// ResponseField represents a single key-value pair of a query.
type ResponseField struct {
	Name  ResponseFieldName
	Value string
}

func newResponseFieldList() ResponseFieldList {
	return make([]ResponseField, 0)
}

// Size returns the number of values stored in this list.
func (l *ResponseFieldList) Size() int {
	return len(*l)
}

// Add adds a sequence of values for the given field name.
func (l *ResponseFieldList) Add(fieldName ResponseFieldName, values ...string) {
	fieldName = fieldName.Normalize()
	for _, val := range values {
		*l = append(*l, ResponseField{fieldName, val})
	}
}

// RemoveAll removes all values for a given field name
func (l *ResponseFieldList) RemoveAll(fieldName ResponseFieldName) {
	fieldName = fieldName.Normalize()
	newList := make([]ResponseField, 0)
	for _, entry := range *l {
		if entry.Name != fieldName {
			newList = append(newList, entry)
		}
	}
	*l = newList
}

// Values returns all values defined for a field name.
func (l *ResponseFieldList) Values(fieldName ResponseFieldName) []string {
	fieldName = fieldName.Normalize()
	arr := make([]string, 0)
	for _, f := range *l {
		if f.Name == fieldName {
			arr = append(arr, f.Value)
		}
	}
	return arr
}

// FirstValue returns the first field value or an empty string for a field name.
func (l *ResponseFieldList) FirstValue(fieldName ResponseFieldName) string {
	fieldName = fieldName.Normalize()
	for _, f := range *l {
		if f.Name == fieldName {
			return f.Value
		}
	}
	return ""
}
