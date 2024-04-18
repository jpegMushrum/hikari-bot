package dict

type Response interface {
	HasEntries() bool
	RelevantKana() string
	RelevantWord() string
	Kanas() []string
	Words() []string
	RelevantSpeechPart() string
	RelevantDefinition() string
}

/// Do we really need this???
type Dictionary[R Response] interface {
	Search(key string) (R, error)
}
