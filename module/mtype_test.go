package module

import "testing"

var legalTypes = []Type{
	TYPE_DOWNLOADER,
	TYPE_ANALYZER,
	TYPE_PIPELINE,
}

var illegalTypes = []Type{
	Type("OTHER_MODULE_TYPE"),
}

func TestTypeCheck(t *testing.T) {
	if CheckType("", fakeModules[0]) {
		t.Fatal("The module type is invalid, but do not be detected!")
	}

	if CheckType(TYPE_DOWNLOADER, nil) {
		t.Fatal("The module is nil, but do not be detected!")
	}

	for _, mt := range legalTypes {
		matchedModule := defaultFakeModuleMap[mt]
		for _, m := range fakeModules {
			if m.ID() == matchedModule.ID() {
				if !CheckType(mt, m) {
					t.Fatalf("Inconsistent module type: expected: %T, actual: %T", matchedModule, mt)
				}
			} else {
				if CheckType(mt, m) {
					t.Fatalf("The module type %T is not matched, but do not be detected!", mt)
				}
			}
		}
	}
}
