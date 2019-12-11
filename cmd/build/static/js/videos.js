
$(function() {


    let player = videojs('example-video', {
        playbackRates: [0.5, 1, 1.5, 2, 4, 8]
    });
    videojs.options.autoplay = true


    player.play();

    console.log(1);
    (function(window, videojs) {
        var player = window.player = videojs('videojs-playbackrate-adjuster-player', {
            playbackRates: [0.5, 1, 1.5, 2, 4]
        });

    }(window, window.videojs));


    let cameras = [
        "video-1",
        "video-2",
        "video-3",
    ];


    window.videosPlayEvents = {
        'click .play': function (e, value, row, index) {
            playVideo(row);
        },
    }

    let $table = $("#table-videos");
        columns = [{
            title: "Date",
            field: "date",
        }];

    $.each(cameras, function(i, c) {
        console.log(c);
        columns.push(
            {
                title: c,
                field: c,
                formatter: videosCanPlayFormatter,
                events: videosPlayEvents,
            }
        );
    });


    //
    // $(".btn-play-video").click(function(e) {
    //     let video = $(e).data("video");
    //     console.log(video);
    // });
    $table.bootstrapTable({
        columns: columns
    });

    // $table.bootstrapTable();


    function playVideo(video) {
        console.log(video);
    }
});
