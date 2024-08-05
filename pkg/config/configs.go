package config

import "path/filepath"

type JobConfig interface {
	SetWorkPath(path string)
	GetWorkPath() string
}

type Extract struct {
	// WorkPath is an exact path to the working directory all other paths are relative to.
	// Inherits from the parent configuration if unset.
	WorkPath string `yaml:"workPath"`

	// ArticlesPath is the filepath to the pages-articles-multistream
	// dump of Wikipedia. Must be compressed as the associated Index points
	// to specific bytes of the compressed format.
	ArticlesPath string `yaml:"articlesPath"`

	// IndexPath is the filepath to the index corresponding to the dump.
	// Accepts bz2 or uncompressed.
	IndexPath string `yaml:"indexPath"`

	// Namespace is the namespace ID to include articles from.
	Namespaces []int `yaml:"namespaces"`

	// OutPath is the filepath to store the database of extracted articles.
	OutPath string `yaml:"outPath"`
}

var _ JobConfig = &Extract{}

func (cfg *Extract) SetWorkPath(path string) {
	cfg.WorkPath = path
}

func (cfg *Extract) GetWorkPath() string {
	return cfg.WorkPath
}

func (cfg *Extract) GetArticlesPath() string {
	if filepath.IsAbs(cfg.ArticlesPath) {
		return cfg.ArticlesPath
	}

	return filepath.Join(cfg.WorkPath, cfg.ArticlesPath)
}

func (cfg *Extract) GetIndexPath() string {
	if filepath.IsAbs(cfg.IndexPath) {
		return cfg.IndexPath
	}

	return filepath.Join(cfg.WorkPath, cfg.IndexPath)
}

func (cfg *Extract) GetOutPath() string {
	if filepath.IsAbs(cfg.OutPath) {
		return cfg.OutPath
	}

	return filepath.Join(cfg.WorkPath, cfg.OutPath)
}

type Clean struct {
	// WorkPath is an exact path to the working directory all other paths are relative to.
	// Inherits from the parent configuration if unset.
	WorkPath string `yaml:"workPath"`

	// ArticlesPath is the filepath to the extracted Wikipedia articles.
	ArticlesPath string `yaml:"articlesPath"`

	// View if set, is a list of cleaned articles to print to the screen (instead of writing to OutPath).
	View []uint `yaml:"view"`

	// OutPath is the filepath to store the database of cleaned articles.
	// Ignored if View is non-empty.
	OutPath string `yaml:"outPath"`
}

var _ JobConfig = &Clean{}

func (cfg *Clean) SetWorkPath(path string) {
	cfg.WorkPath = path
}

func (cfg *Clean) GetWorkPath() string {
	return cfg.WorkPath
}

func (cfg *Clean) GetArticlesPath() string {
	if filepath.IsAbs(cfg.ArticlesPath) {
		return cfg.ArticlesPath
	}

	return filepath.Join(cfg.WorkPath, cfg.ArticlesPath)
}

func (cfg *Clean) GetOutPath() string {
	if filepath.IsAbs(cfg.OutPath) {
		return cfg.OutPath
	}

	return filepath.Join(cfg.WorkPath, cfg.OutPath)
}
