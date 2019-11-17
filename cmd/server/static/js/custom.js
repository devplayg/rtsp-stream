
$(function() {

    let $form = $("#form-streams");

    $(".btn-stream-list").click(function() {
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

    $(".btn-stream-add").click(function() {
        $.ajax({
            url: "/streams",
            type: "POST",
            data:JSON.stringify($form.serializeObject()),
            datatype: 'json'
        }).done(function(data) {
            console.log(data);
        }).fail(function(xhr, status, errorThrown) {
            console.log(xhr);
        });
    });

    $(".btn-stream-start").click(function() {
        let id = $("input[name=id]", $form).val(),
            url = "/streams/" + id + "/start";
        if (id.length < 1) return;

        $.get(url, function() {
            console.log( "start" );
        }).fail(function() {
            console.log( "error" );
        });
    });

    $(".btn-stream-stop").click(function() {
        let id = $("input[name=id]", $form).val(),
            url = "/streams/" + id + "/stop";
        if (id.length < 1) return;

        $.get(url, function() {
            console.log( "stop" );
        }).fail(function() {
            console.log( "error" );
        });
    });

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
    }
});