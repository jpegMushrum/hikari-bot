package dict

type Response interface {
	HasEntries() bool
	RelevantKana() (string, error)
	RelevantWord() (string, error)
	Kanas() []string
	Words() []string
	RelevantSpeechPart() (string, error)
	RelevantDefinition() (string, error)
}

type Dictionary interface {
	Search(key string) (Response, error)
	NounRepr() string
}
