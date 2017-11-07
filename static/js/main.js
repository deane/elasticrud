$(function(){
  for (var prop in mappings) {
    $("<option />")
      .attr("value", prop)
      .html(prop)
      .appendTo("#index");
  }
  $("#index").change(function() {
    $("#type").html("");
    var types = mappings[this.value];
    for (var i in types) {
    $("<option />")
      .attr("value", types[i])
      .html(types[i])
      .appendTo("#type");
    }
  }).trigger("change");

  var index = "";
  var type = "";
  var id = "";

  $("#get").click(function() {
    index = $("#index")[0].value;
    type = $("#type")[0].value;
    id = $("#id")[0].value;

    $.getJSON("/get/"+index+"/"+type+"/"+id)
      .done(function(data) {
        console.log("data :", data);
        editor.set(data);
      })
      .fail(function(data) {
        alert("Fetch Failed: ", data);
      });

  });
  $("#save").click(function() {
    $.post("/set/"+index+"/"+type+"/"+id, JSON.stringify(editor.get()))
      .done(function(data) {
        console.log("saved. data :", data);
      })
      .fail(function(data) {
        alert("Save Failed: ", data);
      });
  });
});
