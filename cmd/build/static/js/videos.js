// https://docs.videojs.com/docs/api/player.html

$(function() {


    let player = videojs('example-video', {
        playbackRates: [0.5, 1, 1.5, 2, 4, 8]
    });
    videojs.options.autoplay = true;


    // player.play();

    let cameras = [
        "video-1",
        "video-2",
        "video-3",
    ];


    window.videosPlayEvents = {
        'click .play': function (e, val, row, idx) {
            playVideo(row);

            let id = $(e.currentTarget).data("id"),
                url = "/videos/" + id + "/date/" + row.date + "/m3u8";

            player.src({
                "type": "application/x-mpegURL",
                "src": url
                //"techOrder": ['youtube'],
                //"youtube": { "iv_load_policy": 3 }
            });
            // if (poster) vgsPlayer.poster(poster);
            player.play();
            // console.log(url);
            // console.log(id);
            // /videos/1/date/20191204/m3u8
            // console.log(e.data("name"));

        },
    }

    let $table = $("#table-videos");
        columns = [{
            title: "Date",
            field: "date",
        }];

    $.each(cameras, function(i, c) {
        console.log(c);
        columns.push({
                title: c,
                field: c,
                formatter: videosCanPlayFormatter,
                events: videosPlayEvents,
        });
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

    $(".btn-test").click(function(e) {
        player.src({
            "type": "application/x-mpegURL",
            "src": "/videos/1/date/20191204/m3u8"
            //"techOrder": ['youtube'],
            //"youtube": { "iv_load_policy": 3 }
        });
        // if (poster) vgsPlayer.poster(poster);
        player.play();

    });
});
