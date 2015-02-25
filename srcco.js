window.onload = function () {
    // can I do this without javascript?
    var codes = document.querySelectorAll(".code");
    for (var i = 0; i < codes.length; i++) {
        var style = window.getComputedStyle(codes[i].parentElement, null);
        codes[i].style.height = style.getPropertyValue("height");
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
};

function triggerTOC(id) {
    var root = document.querySelector("#" + id + "-toc");
    if (!root.classList.contains("active")) {
        root.classList.add("active");
        root.parentNode.classList.add("active");
        root.parentNode.querySelector(".toc-name").classList.add("active");
        root.querySelector(".node-body").classList.add("active");
    } else {
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
}

function triggerNode(node) {
    var body = node.querySelector(".node-body");
    if (!body.classList.contains("active")) {
        console.log('hi');
        body.classList.add("active");
    } else {
        console.log('bye');
        deactivateAll(body.parentNode);
    }
}
