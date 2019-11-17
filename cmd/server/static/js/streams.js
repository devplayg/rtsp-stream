
$(function() {

    let $form = $("#form-streams"),
        $table = $("#table-streams");

    $(".btn-streams-list").click(function() {
        console.log(3);
        $.ajax({
            url: "/streams",
        }).done(function(list) {
            console.log(list);
            if (list.length > 0) {
                $("input[name=id]", $form).val(list[0].id);
            }
        }).fail(function(xhr, status, errorThrown) {
            console.error(xhr);
        });
    });

    $(".btn-streams-add").click(function() {
        $.ajax({
            url: "/streams",
            type: "POST",
            data:JSON.stringify($form.serializeObject()),
            datatype: 'json'
        }).done(function(data) {
            console.log(data);
            $table.bootstrapTable("refresh");
        }).fail(function(xhr, status, errorThrown) {
            console.log(xhr);
        });
    });

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
    };

    $table.bootstrapTable();

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
            }).fail(function() {
                console.log( "error" );
            });
        };

        this.delete = function() {
            let url = "/streams/" + this.id,
                c = this;

            $.ajax({
                url: url,
                type: "DELETE",
            }).done(function(data) {
                console.log(data);
                // c.table.bootstrapTable("refresh");
                $("#table-streams").bootstrapTable('refresh');
            }).fail(function(xhr, status, errorThrown) {
                console.log(xhr);
            });
        };

    }


});