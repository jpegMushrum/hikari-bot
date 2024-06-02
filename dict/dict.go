package dict

type Response interface {
	HasEntries() bool
	RelevantKana() (string, error)
	RelevantWord() (string, error)
	Kanas() []string
	Words() []string
	RelevantSpeechParts() ([]string, error)
	RelevantDefinition() (string, error)
}

type Dictionary interface {
	Search(key string) (Response, error)
	NounRepr() string
	Repr() string
}
