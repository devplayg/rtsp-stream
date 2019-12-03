package m3u8

type Segment struct {
	SeqId    int64   `json:"id"`
	Duration float64 `json:"d"`
	URI      string  `json:"uri"`
	UnixTime int64   `json:"t"`
}

func NewSegment(seqId int64, duration float64, uri string, unixTime int64) *Segment {
	return &Segment{
		SeqId:    seqId,
		Duration: duration,
		URI:      uri,
		UnixTime: unixTime,
	}
}
