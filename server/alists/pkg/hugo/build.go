package hugo

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/otiai10/copy"
)

func (h HugoHelper) Build() {
	a := h.AlistWriter.GetFilesToPublish()
	b := h.AlistsByUserWriter.GetFilesToPublish()

	fmt.Printf("Build static site for %d lists\n", len(a))
	fmt.Printf("Build static site for %d my lists\n", len(b))

	toPublish := append(a, b...)

	if len(toPublish) == 0 {
		fmt.Println("Nothing to publish")
		h.StopCronJob()
		return
	}

	h.buildSite()
	//uuids := h.getPublishedFiles()
	// TODO whats the downside of this?
	// TODO should I have a publish list per type alist, user?

	// Why copy?
	h.copyToSiteCache()

	if h.SiteCacheFolder == h.PublishDirectory {
		h.StopCronJob()
		return
	}

	// Only remove what we processed, that way any that get added will not be lost (hopefully)
	removeA := h.AlistWriter.GetFilesToClean()
	removeB := h.AlistsByUserWriter.GetFilesToClean()
	toDelete := append(removeA, removeB...)
	h.deleteFiles(toDelete)

}

func (h HugoHelper) buildSite() {
	staticSiteFolder := h.Cwd
	// TODO change this to be dynamic via config
	parts := []string{
		"-verbose",
		fmt.Sprintf(`--environment=%s`, h.Environment),
	}

	cmd := exec.Command("hugo", parts...)
	cmd.Dir = staticSiteFolder
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(string(out))
		log.Fatal(err)
	}
	fmt.Println(string(out))
}

// Copy each file over, including non alist files and directories
func (h HugoHelper) copyToSiteCache() {
	fmt.Println("Copy site")
	siteCacheDir := h.SiteCacheFolder
	destinationDir := h.PublishDirectory

	err := copy.Copy(destinationDir, siteCacheDir)
	fmt.Println(err)
	fmt.Println(destinationDir)
	fmt.Println(siteCacheDir)
}

// deleteBuildFiles
// 	- Remove from content
// 	- Remove from data directory
//	- Remove from hugo publishe directory
func (h HugoHelper) deleteBuildFiles(uuid string) {
	panic("REMOVE")
	//files := h.getBuildFiles(uuid)
	//h.deleteFiles(files)
}

func (h HugoHelper) deleteFiles(files []string) {
	// TODO Create an issue about a command line option to purge lists that are not in the database
	// Assume one day, this will get out of sync.
	for _, path := range files {
		fmt.Printf("Removing %s\n", path)

		err := os.Remove(path)
		if err != nil {
			if !strings.HasSuffix(err.Error(), "no such file or directory") {
				fmt.Println(fmt.Sprintf("Failed to remove %s", err))
			}
		}

	}
}
