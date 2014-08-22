var _ = require("underscore");
var Backbone = require("backbone");

var DataCenterMetrics = require("./DataCenterMetrics");

var STATS = [
    "TotalCpus",
    "TotalMemory",
    "TotalDisk",
    "AllocatedCpus",
    "AllocatedCpusPercent",
    "AllocatedMemory",
    "AllocatedMemoryPercent",
    "AllocatedDisk",
    "AllocatedDiskPercent",
    "UsedCpus",
    "UsedCpusPercent",
    "UsedMemory",
    "UsedMemoryPercent",
    "UsedDisk",
    "UsedDiskPercent",
    "SlackCpus",
    "SlackCpusPercent",
    "SlackMemory",
    "SlackMemoryPercent",
    "SlackDisk",
    "SlackDiskPercent"
  ];

module.exports = Backbone.Collection.extend({

  model: DataCenterMetrics,

  initialize: function(models, options) {
    this.options = options;
  },

  parse: function(response, options) {
    var model = this;
    var d3Data = [];
    _.each(STATS, function (key) {
      var series = _.map(response.cluster, function (snapshot) {
        return {
          x: snapshot["Timestamp"],
          y: snapshot[key]
        };
      });
      d3Data.push({name: key, series: series});
    });
    return d3Data;
  },

  url: "/resources"

});
