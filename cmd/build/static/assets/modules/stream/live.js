$(function() {
    checkLiveCameras(streams);

    function checkLiveCameras(streams) {
        // console.log(streams);
        // $.ajax({
        //     url: "/videos",
        // }).done(function(result) {
            if (streams.length < 1) {
                return;
            }

            updateLiveVideos(streams);
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
                "src": "/videos/" + s.id + "/live/m3u8"
            });
            player.ready(function() {
                player.muted(true);
                player.play();
            });
        });
    }
});