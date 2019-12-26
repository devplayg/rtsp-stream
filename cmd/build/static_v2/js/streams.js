let StreamManager = function () {
    this.id = null;

    this.formAdd = $("#form-streams-add");
    this.formEdit = $("#form-streams-edit");
    this.table = $("#table-streams");
    this.modalAdd = $("#modal-streams-add");
    this.modalEdit = $("#modal-streams-edit");

    this.refreshTable = function() {
        this.table.bootstrapTable("refresh");
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

    this.delete = function(id) {
        let c = this;
        $.ajax({
            url: "/streams/" + id,
            type: "DELETE",
        }).done(function(data) {
            c.refreshTable();
        }).fail(function(xhr, status, errorThrown) {
            console.error(xhr);
        });
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
};


let manager = new StreamManager();



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
        // $(".alert", $form).addClass("hide").removeClass("in");
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

let $table = $("#table-streams");



$(".btn-streams-add").click(function() {
    manager.add();
});

$(".btn-streams-update").click(function() {
    manager.update();
});



window.streamsActiveEvents = {
    'click .delete': function (e, value, row, index) {
        manager.delete(row.id);
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
