wsframe.controllers.auth = function (global) {
    "use strict";

    function enter() {
        var type = gg(".form-toggler").html() === "Login" ? "register" : "login";
        var alias = gg(".form-input[name='alias']").attr("value");
        var passhash = sjcl.codec.hex.fromBits(sjcl.hash.sha256.hash(gg(".form-input[name='password']").attr("value")));
        var passhash_repeat = sjcl.codec.hex.fromBits(sjcl.hash.sha256.hash(gg(".form-input[name='password-repeat']").attr("value")));
        var xhr = new XMLHttpRequest();
        var fd = new FormData();

        gg(".form-input[name='password']").attr("value", "");
        gg(".form-input[name='password-repeat']").attr("value", "");
        if (type === "register" && passhash !== passhash_repeat) {
            return;
        }
        fd.append("alias", alias);
        fd.append("passhash", passhash);
        xhr.onload = function () {
            if (xhr.readyState === 4 && xhr.status === 200) {
                global.location.reload();
            }
        };
        xhr.open("POST", type === "login" ? "/login" : "/register", true);
        xhr.setRequestHeader("X-Requested-With", "XMLHttpRequest");
        xhr.responseType = "text";
        xhr.send(fd);
    }
    gg("button[name='enter']").on("click", enter, false);

    function leave() {
        var xhr = new XMLHttpRequest();

        xhr.onload = function () {
            if (xhr.readyState === 4 && xhr.status === 200) {
                global.location.reload();
            }
        };
        xhr.open("POST", "/logout", true);
        xhr.setRequestHeader("X-Requested-With", "XMLHttpRequest");
        xhr.responseType = "text";
        xhr.send(null);
    }
    gg("button[name='leave']").on("click", leave, false);

    function toggleForm(e, node) {
        if (node.text() === "Register") {
            node.html("Login");
            gg(".form-input[name='password-repeat']").remClass("collapsed");
        } else {
            node.html("Register");
            gg(".form-input[name='password-repeat']").addClass("collapsed");
        }
    }
    gg(".form-toggler").on("click", toggleForm, false);

};
