$(function() {
    checkLiveCameras();

    function checkLiveCameras() {
        $.ajax({
            url: prefix+"/videos",
        }).done(function(result) {
            if (result.streams.length < 1) {
                return;
            }

            updateLiveVideos(result.streams);
        });
    }

    function updateLiveVideos(streams) {
        $.each(streams, function(i, s) {
            console.log(s);
            let player = videojs('live'+s.id, {
                controls: true,
                autoplay: false,
                preload: 'auto',
                // liveui: true,
            });

            player.src({
                "type": "application/x-mpegURL",
                "src": prefix+"/videos/" + s.id + "/live/m3u8"
            });
            player.ready(function() {
                player.muted(true);
                player.play();
            });
        });
    }
});