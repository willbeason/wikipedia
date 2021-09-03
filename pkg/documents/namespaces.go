package documents

type Namespace int16

const (
	NamespaceArticle Namespace = iota
	NamespaceTalk
	NamespaceUser
	NamespaceUserTalk
	NamespaceWikipedia
	NamespaceWikipediaTalk
	NamespaceFile
	NamespaceFileTalk
	NamespaceMediaWiki
	NamespaceMediaWikiTalk
	NamespaceTemplate
	NamespaceTemplateTalk
	NamespaceHelp
	NamespaceHelpTalk
	NamespaceCategory
	NamespaceCategoryTalk
	NamespacePortal = 100
	NamespacePortalTalk = 101
	NamespaceDraft = 118
	NamespaceDraftTalk = 119
	NamespaceTimedText = 710
	NamespaceTimedTextTalk = 711
	NamespaceModule = 828
	NamespaceModuleTalk = 829
)
