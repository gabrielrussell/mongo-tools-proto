package mongodump

import (
	"bufio"
	"fmt"
	"github.com/shelman/mongo-tools-proto/common/db"
	commonopts "github.com/shelman/mongo-tools-proto/common/options"
	"labix.org/v2/mgo/bson"
	"os"
)

type MongoDump struct {
	// basic mongo tool options
	ToolOptions *commonopts.ToolOptions

	SessionProvider *db.SessionProvider
}

func (dmp *MongoDump) Dump() error {
	//TODO -- call proper things, track changes

	session := dmp.SessionProvider.GetSession()
	out, err := os.Create("output.bson")
	if err != nil {
		return err
	}
	w := bufio.NewWriter(out)

	collection := session.DB(dmp.ToolOptions.Namespace.DB).C(dmp.ToolOptions.Namespace.Collection)

	fmt.Println(dmp.ToolOptions.Namespace.DB, dmp.ToolOptions.Namespace.Collection)

	cursor := collection.Find(bson.M{}).Iter()
	defer cursor.Close()

	buffChan := make(chan []byte)
	go func() {
		for {
			raw := &bson.Raw{}
			next := cursor.Next(raw)
			if !next {
				close(buffChan)
				return
			}
			buffChan <- raw.Data
		}
	}()

	for {
		buff, alive := <-buffChan
		if alive == false {
			break
		}
		w.Write(buff)
	}
	w.Flush()

	return nil

}
