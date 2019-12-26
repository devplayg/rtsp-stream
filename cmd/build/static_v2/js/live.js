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
            console.log(s.id + " => " + s.status);
            let videoId = 'live'+s.id;

            if (s.status === Started) {
                let player = videojs(videoId, {
                    controls: true,
                    autoplay: false,
                    preload: 'auto',
                    sources: [{
                        type: "application/x-mpegURL",
                        src: "/videos/" + s.id + "/live/m3u8",
                    }]
                });

                // player.src({
                //     type: "application/x-mpegURL",
                //     src: "/videos/" + s.id + "/live/m3u8",
                // });
                player.ready(function() {
                    player.muted(true);
                    player.play();
                });

                return true;
            }

            // player.src({
            //     poster: "/static/img/html5-video.png"
            // });
            let player = videojs(videoId, {
                poster: "/static/img/html5-video.png",
            });
            // player.poster("/static/img/html5-video.png");
        });
    }
});


/*
var player = videojs('#example_video_1', {}, function() {

          this.on('ended', function() {
              player.exitFullscreen();
              player.hasStarted(false);
          });
      });
 */