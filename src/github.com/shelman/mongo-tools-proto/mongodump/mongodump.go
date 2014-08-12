package mongodump

import (
	"bufio"
	"fmt"
	"github.com/shelman/mongo-tools-proto/common/db"
	"github.com/shelman/mongo-tools-proto/common/log"
	commonopts "github.com/shelman/mongo-tools-proto/common/options"
	"github.com/shelman/mongo-tools-proto/common/util"
	"github.com/shelman/mongo-tools-proto/mongodump/options"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

type MongoDump struct {
	// basic mongo tool options
	ToolOptions *commonopts.ToolOptions

	InputOptions  *options.InputOptions
	OutputOptions *options.OutputOptions

	SessionProvider *db.SessionProvider

	// useful internals that we don't directly expose as options
	useStdout bool
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

	dmp.DumpCollection(dmp.ToolOptions.DB, dmp.ToolOptions.Collection)

	return nil
}

func (dmp *MongoDump) DumpCollection(db, c string) {
	if dmp.useStdout {
		//XXX
	} else {
		dbFolder := filepath.Join(dmp.OutputOptions.Out, db)
		err := os.MkdirAll(dbFolder, 0666) //TODO const?
		if err != nil {
			util.Exitf(1, "Error creating directory `%v`: %v", dbFolder, err)
		}

		outFilepath := filepath.Join(dbFolder, fmt.Sprintf("%v.bson", c))
		out, err := os.Create(outFilepath)
		if err != nil {
			util.Exitf(1, "Error creating bson file `%v`: %v", outFilepath, err)
		}
		defer out.Close()

		log.Logf(0, "\t%v.%v to %v", db, c, outFilepath)
		dmp.dumpCollectionToWriter(db, c, bufio.NewWriter(out))

	}

	//TODO metadata
}

func (dmp *MongoDump) dumpCollectionToWriter(db, c string, w *bufio.Writer) {
	session := dmp.SessionProvider.GetSession()

	collection := session.DB(db).C(c)

	log.Logf(1, "Dumping %v.%v", db, c)

	cursor := collection.Find(bson.M{}).Iter()
	defer cursor.Close()

	buffChan := make(chan []byte)
	go func() {
		for {
			raw := &bson.Raw{}
			if err := cursor.Err(); err != nil {
				log.Logf(0, "Error reading from %v.%v: %v", db, c, err)
			}
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
		_, err := w.Write(buff)
		if err != nil {
			log.Logf(0, "Error writing to file: %v", err)
		}
	}
	err := w.Flush()
	if err != nil {
		log.Logf(0, "Error flushing buffered writer...")
	}
}
