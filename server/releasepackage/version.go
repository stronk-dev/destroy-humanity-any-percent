package releasepackage

import "strings"

func CompareReleaseVersions(left, right string) (int, error) {
	leftCore, leftPre, leftOK := splitReleaseVersion(left)
	rightCore, rightPre, rightOK := splitReleaseVersion(right)
	if !leftOK || !rightOK {
		return 0, ErrInvalidContent
	}
	for index := range leftCore {
		if compared := compareNumericIdentifier(leftCore[index], rightCore[index]); compared != 0 {
			return compared, nil
		}
	}
	if len(leftPre) == 0 && len(rightPre) == 0 {
		return 0, nil
	}
	if len(leftPre) == 0 {
		return 1, nil
	}
	if len(rightPre) == 0 {
		return -1, nil
	}
	limit := len(leftPre)
	if len(rightPre) < limit {
		limit = len(rightPre)
	}
	for index := 0; index < limit; index++ {
		leftNumeric, rightNumeric := numericIdentifier(leftPre[index]), numericIdentifier(rightPre[index])
		var compared int
		switch {
		case leftNumeric && rightNumeric:
			compared = compareNumericIdentifier(leftPre[index], rightPre[index])
		case leftNumeric:
			compared = -1
		case rightNumeric:
			compared = 1
		case leftPre[index] < rightPre[index]:
			compared = -1
		case leftPre[index] > rightPre[index]:
			compared = 1
		}
		if compared != 0 {
			return compared, nil
		}
	}
	if len(leftPre) < len(rightPre) {
		return -1, nil
	}
	if len(leftPre) > len(rightPre) {
		return 1, nil
	}
	return 0, nil
}

func validReleaseVersion(value string) bool {
	_, _, valid := splitReleaseVersion(value)
	return valid
}

func splitReleaseVersion(value string) ([3]string, []string, bool) {
	var core [3]string
	parts := strings.SplitN(value, "-", 2)
	coreParts := strings.Split(parts[0], ".")
	if len(coreParts) != len(core) {
		return core, nil, false
	}
	for index := range core {
		if !canonicalNumericIdentifier(coreParts[index]) {
			return core, nil, false
		}
		core[index] = coreParts[index]
	}
	if len(parts) == 1 {
		return core, nil, true
	}
	prerelease := strings.Split(parts[1], ".")
	for _, identifier := range prerelease {
		if identifier == "" || numericIdentifier(identifier) && !canonicalNumericIdentifier(identifier) {
			return core, nil, false
		}
		for _, character := range identifier {
			if character != '-' && (character < '0' || character > '9') && (character < 'A' || character > 'Z') && (character < 'a' || character > 'z') {
				return core, nil, false
			}
		}
	}
	return core, prerelease, true
}

func canonicalNumericIdentifier(value string) bool {
	return numericIdentifier(value) && (value == "0" || value[0] != '0')
}

func numericIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func compareNumericIdentifier(left, right string) int {
	if len(left) < len(right) {
		return -1
	}
	if len(left) > len(right) {
		return 1
	}
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
