package co

// ***************************************************************
// ********************** Chain Operations ***********************
// ***************************************************************
// func (s *SequenceableData[R]) Filter(fn func(R) bool) *SequenceableData[R] {
// 	filteredSlice := make([]*executable[R], 0, len(s.executables))
// 	for i, executable := range s.executables {
// 		if executable.Error != nil {
// 			filteredSlice = append(filteredSlice, s.executables[i])
// 			continue
// 		}
// 		if fn(executable.value) {
// 			filteredSlice = append(filteredSlice, s.executables[i])
// 		}
// 	}

// 	s.executables = filteredSlice
// 	return s
// }

// func (s *SequenceableData[R]) Compacted() *SequenceableData[R] {
// 	filteredSlice := make([]*executable[R], 0, len(s.executables))
// 	for i, executable := range s.executables {
// 		if executable.Error != nil {
// 			continue
// 		}
// 		if reflect.DeepEqual(s.executables[i].value, *new(R)) {
// 			continue
// 		}

// 		filteredSlice = append(filteredSlice, s.executables[i])
// 	}

// 	s.executables = filteredSlice
// 	return s
// }

// func (s *SequenceableData[R]) RemoveDuplicates() *SequenceableData[R] {
// 	filteredSlice := make([]*executable[R], 0, len(s.executables))
// 	for i, executable := range s.executables {
// 		if executable.Error != nil {
// 			continue
// 		}

// 		hasDuplicates := false
// 		for j := range filteredSlice {
// 			if reflect.DeepEqual(s.executables[i].value, filteredSlice[j].value) {
// 				hasDuplicates = true
// 				break
// 			}
// 		}
// 		if hasDuplicates {
// 			continue
// 		}

// 		filteredSlice = append(filteredSlice, s.executables[i])
// 	}

// 	s.executables = filteredSlice
// 	return s
// }

// func (s *SequenceableData[R]) RemoveDuplicatesBy(fn func(R, R) bool) *SequenceableData[R] {
// 	filteredSlice := make([]*executable[R], 0, len(s.executables))
// 	for i, executable := range s.executables {
// 		if executable.Error != nil {
// 			continue
// 		}

// 		hasDuplicates := false
// 		for j := range filteredSlice {
// 			if fn(s.executables[i].value, filteredSlice[j].value) {
// 				hasDuplicates = true
// 				break
// 			}
// 		}
// 		if hasDuplicates {
// 			continue
// 		}

// 		filteredSlice = append(filteredSlice, s.executables[i])
// 	}

// 	s.executables = filteredSlice
// 	return s
// }
