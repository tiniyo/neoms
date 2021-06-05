package tinixml

import (
	"time"
	"strconv"
	"github.com/beevik/etree"
	"github.com/neoms/logger"
)

/*
	<Response>
    	<Pause length="8"></Pause>
	</Response>
*/

func ProcessPause(uuid string, element *etree.Element) {
	pauseTime := time.Duration(1)
	for _, attr := range element.Attr {
                logger.Logger.Debug("ATTR: %s=%s\n", attr.Key, attr.Value)
                if attr.Key == "length"{
			pauseT, err  := strconv.Atoi(attr.Value)
                        if pauseT == 0 || err != nil{
                                pauseTime = time.Duration(1)
                        }else{
				pauseTime = time.Duration(pauseT)
			}
                }
        }
	time.Sleep(pauseTime * time.Second)
	return
}


func ProcessPauseTime(dur int) {
	pauseTime := time.Duration(dur)
	time.Sleep(pauseTime * time.Second)
	return
}
