$(function () {
    let VideoShop = function () {
        this.table = $("#table-videos");
        this.columns = [];
        this.player = videojs('player', {
            controls: true,
            autoplay: false,
            preload: 'auto',
            playbackRates: [0.5, 1, 1.5, 2, 4, 8],
        });

        this.initTable = function () {
            let c = this;

            $.ajax({
                url: "/streams",
            }).done(function (streams) {
                if (streams.length < 1) {
                    return;
                }

                // console.log(streams);
                c.columns.push({
                    title: "Date",
                    field: "date",
                    formatter: videosDateFormatter,
                });

                $.each(streams, function (i, s) {
                    c.columns.push({
                        title: s.name,
                        field: "video-" + s.id,
                        formatter: videosVideoFormatter,
                        events: videosPlayEvents,
                    });
                });

                c.table.bootstrapTable({
                    columns: c.columns,
                });
            });
        };

        // this.updateVideos = function () {
            // this.val = 3;
            // console.log({
            //     columns: [{
            //         title: 'ID',
            //         field: 'id'
            //     }, {
            //         title: 'Item Name',
            //         field: 'name'
            //     }, {
            //         title: 'Item Price',
            //         field: 'price'
            //     }]
            // });
            //
            // console.log(this.columns);
            //
            //
            // this.table.bootstrapTable({
            //     columns: this.columns,
            // });
            //     $table.bootstrapTable({
            //         columns: columns,
            //     });
        // };

        this.play = function (uri) {
            this.player.src({
                "type": "application/x-mpegURL",
                "src": uri
            });
            this.player.ready(function () {
                player.play();
            });

            $(".example-modal-fullscreen").modal("show");
        };

        this.init = function () {
            this.initTable();
        };

        this.init();
    };

    shop = new VideoShop();

    window.videosPlayEvents = {
        'click .video': function (e, val, row, idx) {
            let id = $(e.currentTarget).data("id"),
                url = "/videos/" + id + "/date/" + row.date + "/m3u8";
            shop.play(url);
        },
        'click .live': function (e, val, row, idx) {
            let id = $(e.currentTarget).data("id"),
                url = "/videos/" + id + "/live/m3u8";
            shop.play(url);
        },
        'click .today': function (e, val, row, idx) {
            let id = $(e.currentTarget).data("id"),
                url = "/videos/" + id + "/today/m3u8";
            shop.play(url);
        },
    };


    // new VideoShop
    // let $table = $("#table-videos"),
    //     player = videojs('player', {
    //         controls: true,
    //         autoplay: false,
    //         preload: 'auto',
    //         playbackRates: [0.5, 1, 1.5, 2, 4, 8],
    //     });
    //
    // initTable();
    // // updateVideos();
    // // $table.bootstrapTable("load", result.videos);
    //
    // function updateVideos() {
    //     $.ajax({
    //         url: "/streams",
    //     }).done(function(streams) {
    //         console.log(streams);
    //         if (streams.length < 1) {
    //             return;
    //         }
    //
    //         updateTable(streams)
    //     });
    // }
    //
    // function initTable() {
    //     let columns = [{
    //         title: "Date",
    //         field: "date",
    //     }];
    //
    //     $.each(result.streams, function(i, s) {
    //         console.log(s);
    //         columns.push({
    //             title: "Camera-" + s.id,
    //             field: "video-" + s.id,
    //             formatter: videosCanPlayFormatter,
    //             events: videosPlayEvents,
    //         });
    //     });
    //
    //     $table.bootstrapTable({
    //         columns: columns,
    //     });
    //
    // }
    //
    // window.videosPlayEvents = {
    //     'click .play': function (e, val, row, idx) {
    //         let id = $(e.currentTarget).data("id"),
    //             url = "/videos/" + id + "/date/" + row.date + "/m3u8";
    //         playVideo(url);
    //     },
    //     'click .live': function (e, val, row, idx) {
    //         let id = $(e.currentTarget).data("id"),
    //             url = "/videos/" + id + "/live/m3u8";
    //         playVideo(url);
    //     },
    //     'click .today': function (e, val, row, idx) {
    //         let id = $(e.currentTarget).data("id"),
    //             url = "/videos/" + id + "/today/m3u8";
    //         playVideo(url);
    //     },
    // };
    //
    // function playVideo(uri, live) {
    //     player.src({
    //         "type": "application/x-mpegURL",
    //         "src": uri
    //     });
    //     player.ready(function () {
    //         player.play();
    //     });
    //
    //     // data-toggle="modal" data-target=".example-modal-fullscreen"
    //     $(".example-modal-fullscreen").modal("show");
    // }

});