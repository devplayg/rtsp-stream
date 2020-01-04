let StreamManager = function () {
    this.id = null;

    this.formAdd = $("#form-streams-add");
    this.formEdit = $("#form-streams-edit");
    this.table = $("#table-streams");
    this.player = videojs('player', {
        controls: true,
        autoplay: false,
        preload: 'auto',
        // playbackRates: [0.5, 1, 1.5, 2, 4, 8],
    });

    // Modal
    this.modalAdd = $("#modal-streams-add");
    this.modalEdit = $("#modal-streams-edit");
    this.modalPlayer = $("modal-streams-player");


    this.refreshTable = function(silent) {
        if (silent === undefined) {
            silent = false;
        }
        this.table.bootstrapTable("refresh", {silent: silent});
    };

    this.add = function() {
        let data = this.formAdd.serializeObject(),
            c = this;
        $.ajax({
            url: "/streams",
            method: "POST",
            data: JSON.stringify(data),
            dataType: "json",
        }).done(function(data) {
            console.log(data);
            c.modalAdd.modal("hide");
            c.refreshTable();
        }).fail(function(xhr, status, errorThrown) {
            console.log(xhr);
            c.formAdd.find(".alert .msg").text(xhr.responseJSON.error);
            c.formAdd.find(".alert").removeClass("d-none");
        });
    };

    this.update = function() {
        let data = this.formEdit.serializeObject(),
            c = this;
        $.ajax({
            url: "/streams/" + this.id,
            method: "PATCH",
            data: JSON.stringify(data),
            dataType: "json",
        }).done(function(data) {
            // console.log(data);
            c.modalEdit.modal("hide");
        }).fail(function(xhr, status, errorThrown) {
            console.log(xhr);
            c.formEdit.find(".alert .msg").text(xhr.responseJSON.error);
            c.formEdit.find(".alert").removeClass("d-none");
        });
    };

    this.start = function(id) {
        let c = this;
        $.get("/streams/" + id + "/start", function() {
            c.refreshTable();
        }).fail(function(xhr) {
            console.error(xhr);
        });
    };

    this.stop = function(id) {
        let c = this;
        $.get("/streams/" + id + "/stop", function() {
            c.refreshTable();
        }).fail(function(xhr) {
            console.error(xhr);
        });
    };

    this.delete = function(row) {
        let c = this;
        Swal.fire({
            title: 'Are you sure you want to delete stream ?',
            text: row.name,
            type: "warning",
            showCancelButton: true,
            // confirmButtonColor: '#3085d6',
            // cancelButtonColor: '#d33',
            confirmButtonText: 'Yes, delete it!'
        }).then((result) => {
            if (result.value) {
                $.ajax({
                    url: "/streams/" + row.id,
                    type: "DELETE",
                }).done(function(data) {
                    c.refreshTable();
                    Swal.fire(
                        'Deleted!',
                        'Your file has been deleted.',
                        'success'
                    )
                }).fail(function(xhr, status, errorThrown) {
                    console.error(xhr);
                    Swal.fire(
                        'Fail',
                        xhr.responseJSON.error,
                        'error'
                    )
                });
            }
        })

    };

    this.show = function(id) {
        this.id = id;

        let $form = $("#form-streams-edit"),
            c = this;
        $.ajax({
            url: "/streams/" + id,
        }).done(function(stream) {
            console.log(stream);
            $("input[name=id]", $form).val(stream.id);
            $("input[name=name]", $form).val(stream.name);
            $("input[name=uri]", $form).val(stream.uri);
            $("input[name=username]", $form).val(stream.username);
            $("input[name=enabled]", $form).prop("checked", stream.enabled);
            $("input[name=recording]", $form).prop("checked", stream.recording);

            c.modalEdit.modal("show");

        }).fail(function(xhr, status, errorThrown) {
            console.log(xhr);
            $form.find(".alert .msg").text(xhr.responseJSON.error);
        });
    };

    this.playVideo = function (uri) {
        this.player.src({
            "type": "application/x-mpegURL",
            "src": uri
        });
        let c = this;
        this.player.ready(function () {
            c.player.muted(true);
            c.player.play();
        });
        $("#modal-streams-player").modal("show");
    };

    this.stopVideo = function() {
        // $("#player").pause();
        // this.player.currentTime = 0;
        $(".video-js")[0].player.pause();
    };
};


let manager = new StreamManager();
setInterval(function() {
    manager.refreshTable(true);
}, 3000);



$.fn.serializeObject = function() {
    let result = {};
    let extend = function(i, e) {
        let node = result[e.name];
        if (typeof node !== "undefined" && node !== null) {
            if ($.isArray(node)) {
                node.push(e.value);
            } else {
                result[e.name] = [node, e.value];
            }
        } else {
            result[e.name] = e.value;
        }

        if ($("input[name=" + e.name + "]").attr('type') === "checkbox") {
            console.log(e.name);
            if (e.value === "on") {
                result[e.name] = true;
            } else {
                result[e.name] = false;
            }
        }
    };

    $.each(this.serializeArray(), extend);
    return result;
};

$(".modal-form")
    .on("hidden.bs.modal", function () {
        let $form = $(this).closest("form");
        // $form.validate().resetForm();
        $form.get(0).reset();
        $(".alert", $form).addClass("d-none");
        // $(".alert .message", $form).empty();

        manager.refreshTable();
    })
    .on("shown.bs.modal", function () {
        let $form = $(this).closest("form");
        $form.find("input:not(readonly)[type=text],input[type=password],textarea")
            .filter(":visible:first")
            .focus()
            .select();
    });


$("#modal-streams-player")
    .on("hidden.bs.modal", function () {
        console.log(11);
        manager.stopVideo();
    });

let $table = $("#table-streams");



$(".btn-streams-add").click(function() {
    manager.add();
});

$(".btn-streams-update").click(function() {
    manager.update();
});



window.streamsActiveEvents = {
    'click .delete': function (e, value, row, index) {
        manager.delete(row);
    },
    'click .start': function (e, value, row, index) {
        manager.start(row.id);
    },
    'click .stop': function (e, value, row, index) {
        manager.stop(row.id);
    },
    'click .edit': function (e, value, row, index) {
        manager.show(row.id);
    },
};

window.streamsStatusEvents = {
    'click .live': function (e, val, row, idx) {
        let url = "/live/" + row.id + "/m3u8";
        console.log(url);
        manager.playVideo(url);
    },
};
