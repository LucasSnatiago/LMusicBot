package music

import (
	"bytes"
	"os"
)

func GetSongByteArray(search string) *bytes.Buffer {
	// yt := ytdlp.New().
	// 	YesPlaylist().
	// 	Format("bestaudio[ext=webm]").
	// 	DefaultSearch("ytsearch").
	// 	Output("teste.mkv")

	// _, err := yt.Run(context.TODO(), search)
	// if err != nil {
	// 	log.Printf("Failed to find %s, %s", search, err)
	// }

	file, _ := os.ReadFile("teste.mkv")

	return bytes.NewBuffer(file)
}
