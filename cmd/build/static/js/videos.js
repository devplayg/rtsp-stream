// https://docs.videojs.com/docs/api/player.html

$(function() {

    let player = videojs('example-video', {
        playbackRates: [0.5, 1, 1.5, 2],
    });


    // videojs.options.autoplay = true;
    let $table = $("#table-videos");

    // let player = videojs.getPlayer("example-video").ready(function(){
        // // When the player is ready, get a reference to it
        // var myPlayer = this;
        // // +++ Define the playback rate options +++
        // var options = {"playbackRates":[0.5, 1, 1.5, 2, 4]};
        // // +++ Initialize the playback rate button +++
        // myPlayer.controlBar.playbackRateMenuButton = myPlayer.controlBar.addChild('PlaybackRateMenuButton', {
        //     playbackRates: options.playbackRates
        // });
    // });

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
            let id = $(e.currentTarget).data("id"),
                url = "/videos/" + id + "/date/" + row.date + "/m3u8";

            playVideo(url, false);
        },
        'click .live': function (e, val, row, idx) {
            let id = $(e.currentTarget).data("id"),
                url = "/videos/" + id + "/live/m3u8";
            playVideo(url, true);
        },
        'click .today': function (e, val, row, idx) {
            let id = $(e.currentTarget).data("id"),
                url = "/videos/" + id + "/today/m3u8";
            playVideo(url, false);
        },
    };

    function playVideo(src, live) {
        player.src({
            "type": "application/x-mpegURL",
            "src": src
        });
        player.play();
        // videojs.getPlayer("example-video").src({
        //     "type": "application/x-mpegURL",
        //     "src": src
        // });

        // if (live) {
        //     videojs.getPlayer("example-video").controlBar.removeChild("PlaybackRateMenuButton");
        // } else {
        //     videojs.getPlayer("example-video").controlBar.addChild('PlaybackRateMenuButton', {
        //         playbackRates:  [0.5, 1, 1.5, 2, 4],
        //     });
        // }

        // When the player is ready, get a reference to it
        // var myPlayer = this;
        // // +++ Define the playback rate options +++
        // var options = {"playbackRates":[0.5, 1, 1.5, 2, 4]};
        // // +++ Initialize the playback rate button +++
        // myPlayer.controlBar.playbackRateMenuButton = myPlayer.controlBar.addChild('PlaybackRateMenuButton', {
        //     playbackRates: options.playbackRates
        // });

        // videojs.getPlayer("example-video").play();





        // videojs.getPlayer("example-video").ready(function(){
        //     // When the player is ready, get a reference to it
        //     var myPlayer = this;
        //     options = {"playbackRates":[0.5, 1, 1.5, 2, 4]};
        //     myPlayer.controlBar.playbackRateMenuButton = myPlayer.controlBar.addChild('PlaybackRateMenuButton', {
        //         playbackRates: options.playbackRates
        //     });
        // });

    }
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
