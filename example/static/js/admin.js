"use strict";

import gg from "./gg.js";
import wsframe from "./wsframe.js";

wsframe.controllers.admin = function admin(global) {
    "use strict";

    gg(".index").insert("beforeend", " admin");
};
