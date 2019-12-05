package streaming

import "os/exec"

func GetHlsStreamingCommand(stream *Stream) *exec.Cmd {
	return exec.Command(
		"ffmpeg",
		"-y",
		"-fflags",
		"nobuffer",
		"-rtsp_transport",
		"tcp",
		"-i",
		stream.StreamUri(),
		"-vsync",
		"0",
		"-copyts",
		"-vcodec",
		"copy",
		"-movflags",
		"frag_keyframe+empty_moov",
		"-an",
		"-hls_flags",
		"append_list",
		"-f",
		"hls",
		"-segment_list_flags",
		"live",
		"-hls_time",
		"1",
		"-hls_list_size",
		"3",
		//"-hls_time",
		//"60",
		"-hls_segment_filename",
		stream.liveDir+"/"+stream.ProtocolInfo.LiveFilePrefix+"%d.ts",
		stream.liveDir+"/"+stream.ProtocolInfo.MetaFileName,
	)
	//output, err := cmd.CombinedOutput()
	//if err != nil {
	//    log.Error(string(output))
	//    return err
	//}
}
