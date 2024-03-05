package documents

// A Namespace is a page property that refers to its function within Wikipedia.
// pages with the same Namespace have the same structure and function, and can usually be processed together.
type Namespace int16

const (
	// NamespaceArticle applies to pages containing written content of Wikipedia: the text itself and what people
	// generally mean when they say "Wikipedia".
	NamespaceArticle Namespace = iota
	// NamespaceTalk pages have a corresponding Article, and are where editors discuss the content of the Article.
	NamespaceTalk
	NamespaceUser
	NamespaceUserTalk
	// NamespaceWikipedia pages are the procedural guidelines of how Wikipedia is run and maintained, such as style
	// guides.
	NamespaceWikipedia
	// NamespaceWikipediaTalk pages are where editors discuss how Wikipedia is run and propose changes.
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
	NamespacePortal        = 100
	NamespacePortalTalk    = 101
	NamespaceDraft         = 118
	NamespaceDraftTalk     = 119
	NamespaceTimedText     = 710
	NamespaceTimedTextTalk = 711
	NamespaceModule        = 828
	NamespaceModuleTalk    = 829
	NamespaceMedia         = -2
	NamespaceSpecial       = -1
)
