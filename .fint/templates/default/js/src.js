function toggleMessages(id) {
  var elem = document.getElementById(id);
  if (elem.style.display == "none") {
    elem.style.display = "table-row";
  } else {
    elem.style.display = "none";
  }
}
