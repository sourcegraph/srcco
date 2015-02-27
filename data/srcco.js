window.onload = function () {
    // This is a dirty hack so that we don't expand the code boxes if
    // we think the user is on a phone when the window has finished
    // loading. There should be a way to do this "responsively".
    if (!window.matchMedia('(max-device-width: 600px)').matches) {
        var codes = document.querySelectorAll(".code");
        for (var i = 0; i < codes.length; i++) {
            var style = window.getComputedStyle(codes[i].parentElement, null);
            codes[i].style.height = style.getPropertyValue("height");
        }
    }
    var allTOCs = document.querySelectorAll(".toc");
    for (var i = 0; i < allTOCs.length; i++) {
        adjustIndent(allTOCs[i]);
        var parent = allTOCs[i].parentNode;
        var fullHeight = window.innerHeight - (parent.offsetTop + parent.offsetHeight);
        allTOCs[i].style.height =  fullHeight + "px";
    }
    var names = document.querySelectorAll(".toc-name");
    for (var i = 0; i < names.length; i++) {
        names.item(i).addEventListener("click", function(ev) {
            triggerTOC(ev.target.id);
        });
    }
    var titles = document.querySelectorAll(".node-title");
    for (var i = 0; i < titles.length; i++) {
        titles.item(i).addEventListener("click", function(ev) {
            triggerNode(ev.target.parentNode);
        });
    }
    document.addEventListener("click", function(ev) {
        if (closestClass(ev.target, "tocs") === undefined) {
            for (var i = 0; i < allTOCs.length; i++) {
                deactivateTOC(allTOCs[i]);
            }
        }
    });
};

function closestClass(elem, className) {
    // Get closest match
    for (; elem && elem !== document; elem = elem.parentNode) {
        if (elem.classList.contains(className)) {
            return elem;
        }
    }
}

function adjustIndent(tocNode) {
    var indentMultiplier = 5;
    var nodes = tocNode.querySelectorAll(".node");
    for (var i = 0; i < nodes.length; i++) {
        // First, we get the indent.
        var level = Number(nodes[i].getAttribute("level"));
        // Then we apply the indent to the node-title at one less than
        // the indent level.
        if (level !== 0) {
            var title = nodes[i].querySelector(".node-title");
            title.style.textIndent = (indentMultiplier * (level - 1)) + "px";
        }
        // Finally, we apply the indent to all of the node-path
        // elements.
        var paths = nodes[i].querySelectorAll(".node-path");
        for (var j = 0; j < paths.length; j++) {
            paths[j].style.textIndent = (indentMultiplier * level) + "px";
        }
    }
}

function triggerTOC(id) {
    var allTOCs = document.querySelectorAll(".toc");
    var tocID = id + "-toc";
    var activated;
    for (var i = 0; i < allTOCs.length; i++) {
        var root = allTOCs[i];
        if (root.id == tocID && !root.classList.contains("active")) {
            root.classList.add("active");
            activated = root;
            root.parentNode.querySelector(".toc-name").classList.add("active");
            root.querySelector(".node-body").classList.add("active");
            root.querySelector(".carrot").classList.add("fa-rotate-90");
        } else {
            deactivateTOC(root)
        }
    }
    if (activated) {
        root.parentNode.classList.add("active");
    }
}

function deactivateTOC(root) {
    if (root.classList.contains("active")) {
        root.classList.remove("active");
        root.parentNode.classList.remove("active");
        root.parentNode.querySelector(".toc-name").classList.remove("active");
        deactivateAll(root.querySelector(".node"));
    }
}

function deactivateAll(node) {
    var nodes = node.querySelectorAll(".node-body");
    for (var i = 0; i < nodes.length; i++) {
        nodes[i].classList.remove("active");
    }
    var carrots = node.querySelectorAll(".carrot");
    for (var i = 0; i < carrots.length; i++) {
        carrots[i].classList.remove("fa-rotate-90");
    }
}

function triggerNode(node) {
    var body = node.querySelector(".node-body");
    if (!body.classList.contains("active")) {
        body.classList.add("active");
        body.parentNode.querySelector(".carrot").classList.add("fa-rotate-90");
    } else {
        deactivateAll(body.parentNode);
    }
}
