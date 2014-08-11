package mongodump

import (
	"bufio"
	"fmt"
	"github.com/shelman/mongo-tools-proto/common/db"
	"github.com/shelman/mongo-tools-proto/common/log"
	commonopts "github.com/shelman/mongo-tools-proto/common/options"
	"github.com/shelman/mongo-tools-proto/mongodump/options"
	"labix.org/v2/mgo/bson"
	"os"
)

type MongoDump struct {
	// basic mongo tool options
	ToolOptions *commonopts.ToolOptions

	InputOptions *options.InputOptions

	SessionProvider *db.SessionProvider
}

func (dmp *MongoDump) ValidateOptions() error {
	switch {
	case dmp.InputOptions.Query != "" && dmp.ToolOptions.Collection != "":
		return fmt.Errorf("cannot dump using a query without a specific collection")
	}
	return nil
}

func (dmp *MongoDump) Dump() error {
	//TODO -- call proper things, track changes

	//TODO move this outside of this file
	if err := dmp.ValidateOptions(); err != nil {
		return err
	}

	session := dmp.SessionProvider.GetSession()
	out, err := os.Create("output.bson")
	if err != nil {
		return err
	}
	w := bufio.NewWriter(out)

	collection := session.DB(dmp.ToolOptions.Namespace.DB).C(dmp.ToolOptions.Namespace.Collection)

	log.Logf(0, "DATABASE %v, %v", dmp.ToolOptions.Namespace.DB, dmp.ToolOptions.Namespace.Collection)

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

func (dmp *MongoDump) DumpCollection(c string) {
	//TODO bson
	//TODO metadata
}
