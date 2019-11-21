
$(function() {

    let $table = $("#table-streams");


    $(".btn-streams-debug").click(function() {
        $.get("/streams/debug", function() {
        });
    });

    $(".btn-streams-add").click(function() {
        console.log("add");
        let $form = $("#form-streams-add"),
            url = "/streams";

        $.ajax({
            url: url,
            method: "POST",
            data: JSON.stringify($form.serializeObject()),
            dataType: "json",
        }).done(function(data) {
            console.log(data);
            $table.bootstrapTable("refresh");
        }).fail(function(xhr, status, errorThrown) {
            console.log(xhr);
            $form.find(".alert .msg").text(xhr.responseJSON.error);
        });
    });

    $(".btn-streams-update").click(function() {
        let $form = $("#form-streams-edit"),
            id = $("input[name=id]", $form).val(),
            url = "/streams/" + id;

        let data = $form.serializeObject();
        data.id =  parseInt(data.id, 10);
        // console.log(data);

        $.ajax({
            url: url,
            method: "PATCH",
            data: JSON.stringify(data),
            dataType: "json",
        }).done(function(result) {
            console.log(result);
            $table.bootstrapTable("refresh");
        }).fail(function(xhr, status, errorThrown) {
            console.log(xhr);
            $form.find(".alert .msg").text(xhr.responseJSON.error);
        });
    });

    // $table.bootstrapTable('getSelections').

    // $(".btn-streams-start").click(function() {
    //     let id = $("input[name=id]", $form).val(),
    //         url = "/streams/" + id + "/start";
    //     if (id.length < 1) return;
    //
    //     $.get(url, function() {
    //         console.log( "start" );
    //     }).fail(function() {
    //         console.log( "error" );
    //     });
    // });
    //
    // $(".btn-streams-stop").click(function() {
    //     let id = $("input[name=id]", $form).val(),
    //         url = "/streams/" + id + "/stop";
    //     if (id.length < 1) return;
    //
    //     $.get(url, function() {
    //         console.log( "stop" );
    //     }).fail(function() {
    //         console.log( "error" );
    //     });
    // });

    $.fn.serializeObject = function() {
        var result = {}
        var extend = function(i, element) {
            var node = result[element.name]
            if ("undefined" !== typeof node && node !== null) {
                if ($.isArray(node)) {
                    node.push(element.value)
                } else {
                    result[element.name] = [node, element.value]
                }
            } else {
                result[element.name] = element.value
            }
        }

        $.each(this.serializeArray(), extend)
        return result
    };

    window.streamsActiveEvents = {
        'click .delete': function (e, value, row, index) {
            let stream = new Stream(row.id);
            stream.delete();
        },
        'click .start': function (e, value, row, index) {
            let stream = new Stream(row.id);
            stream.start();
        },
        'click .stop': function (e, value, row, index) {
            let stream = new Stream(row.id);
            stream.stop();
        },

        'click .edit': function (e, value, row, index) {
            let stream = new Stream(row.id);
            stream.showEdit();
        },
    };

    $table.bootstrapTable();

    // $table.on('click-row.bs.table', function (e, row, $element) {
    //     console.log(row);
    // });

    // $(".btn-streams-active").click(function(e) {
    //     let id = $(this).data("id");
    //     console.log(id);
        // let id = $("input[name=id]", $form).val(),
        //     url = "/streams/" + id + "/start";
        // if (id.length < 1) return;
        //
        // $.get(url, function() {
        //     console.log( "start" );
        // }).fail(function() {
        //     console.log( "error" );
        // });
    // });


    let Stream = function (id, key) {
        this.id = id;

        this.form = $("#form-" + key);

        this.table = $("#table-" + key);

        this.start = function() {
            let c = this;
            $.get("/streams/" + id + "/start", function() {
                console.log( "start:" + c.id );
                c.table.bootstrapTable("refresh");
            }).fail(function() {
                console.log( "error" );
            });
        };

        this.stop = function() {
            let c = this;
            $.get("/streams/" + id + "/stop", function() {
                console.log( "stopped:" + c.id );
                c.table.bootstrapTable("refresh");
            }).fail(function(xhr, status, errorThrown) {
                console.log(xhr);
            });
        };

        this.delete = function() {
            let url = "/streams/" + this.id;

            $.ajax({
                url: url,
                type: "DELETE",
            }).done(function(data) {
                $table.bootstrapTable('refresh');
            }).fail(function(xhr, status, errorThrown) {
                console.log(xhr);
            });
        };

        this.showEdit = function() {
            let url = "/streams/" + this.id;
            let $form = $("#form-streams-edit");
            $.ajax({
                url: url,
            }).done(function(stream) {
                console.log(stream);
                $("input[name=id]", $form).val(stream.id);
                $("input[name=uri]", $form).val(stream.uri);
                $("input[name=username]", $form).val(stream.username);
                $("input[name=password]", $form).val(stream.password);
            }).fail(function(xhr, status, errorThrown) {
                console.log(xhr);
                $form.find(".alert .msg").text(xhr.responseJSON.error);
            });
        };

    }


});

// $(".btn-streams-list").click(function() {
//     let $form = $("#form-streams-add");
//     $.ajax({
//         url: "/streams",
//     }).done(function(list) {
//         console.log(list);
//         if (list.length > 0) {
//             $("input[name=id]", $form).val(list[0].id);
//         }
//     }).fail(function(xhr, status, errorThrown) {
//         console.log(xhr);
//     });
// });
