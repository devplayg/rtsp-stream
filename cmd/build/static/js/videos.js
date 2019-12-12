// https://docs.videojs.com/docs/api/player.html

$(function() {

    let player = videojs('example-video', {
        playbackRates: [0.5, 1, 1.5, 2, 4, 8]
    });
    videojs.options.autoplay = true;
    let $table = $("#table-videos");



    updateVideos();

    function updateVideos() {
        $.ajax({
            url: "/videos",
        }).done(function(result) {
            console.log(result);
            if (result.streams.length < 1) {
                return;
            }

            updateTable(result)
        });
    }

    function updateTable(result) {
        let columns = [{
            title: "Date",
            field: "date",
        }];

        $.each(result.streams, function(i, s) {
            console.log(s);
            columns.push({
                    title: "Camera-" + s.id,
                    field: "video-" + s.id,
                    formatter: videosCanPlayFormatter,
                    events: videosPlayEvents,
            });
        });

        $table.bootstrapTable({
            columns: columns,
        });
        $table.bootstrapTable("load", result.videos);

    }



    // player.play();
    //
    // let cameras = [
    //     "video-1",
    //     "video-2",
    // ];
    //
    //
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
        },
    };
    //
    // let $table = $("#table-videos");
    //     columns = [{
    //         title: "Date",
    //         field: "date",
    //     }];
    //
    // $.each(cameras, function(i, c) {
    //     console.log(c);
    //     columns.push({
    //             title: c,
    //             field: c,
    //             formatter: videosCanPlayFormatter,
    //             events: videosPlayEvents,
    //     });
    // });
    //
    //
    // $table.bootstrapTable({
    //     columns: columns
    // });
    //
    // function playVideo(video) {
    //     console.log(video);
    // }
    //
    // $(".btn-test").click(function(e) {
    //     player.src({
    //         "type": "application/x-mpegURL",
    //         "src": "/videos/1/date/20191204/m3u8"
    //     });
    //     player.play();
    // });
});
