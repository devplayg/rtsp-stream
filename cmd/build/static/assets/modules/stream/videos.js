$(function() {
    let $table = $("#table-videos"),
        player = videojs('player', {
            controls: true,
            autoplay: false,
            preload: 'auto',
            playbackRates: [0.5, 1, 1.5, 2, 4, 8],
        });

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

    window.videosPlayEvents = {
        'click .play': function (e, val, row, idx) {
            let id = $(e.currentTarget).data("id"),
                url = "/videos/" + id + "/date/" + row.date + "/m3u8";
            playVideo(url);
        },
        'click .live': function (e, val, row, idx) {
            let id = $(e.currentTarget).data("id"),
                url = "/videos/" + id + "/live/m3u8";
            playVideo(url);
        },
        'click .today': function (e, val, row, idx) {
            let id = $(e.currentTarget).data("id"),
                url = "/videos/" + id + "/today/m3u8";
            playVideo(url);
        },
    };

    function playVideo(uri, live) {
        player.src({
            "type": "application/x-mpegURL",
            "src": uri
        });
        player.ready(function() {
            player.play();
        });
    }

});