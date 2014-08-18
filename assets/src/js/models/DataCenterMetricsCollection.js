var _ = require("underscore");
var Backbone = require("backbone");

var DataCenterMetrics = require ("./DataCenterMetrics");

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
    "SlackDiskPercent",
    "Timestamp"
  ];

module.exports = Backbone.Collection.extend({

  model: DataCenterMetrics,

  initialize: function(models, options) {
    this.options = options;
  },

  fetch: function (response) {
    var model = this;
    response = [
        {
          "TotalCpus": 1,
          "TotalMemory": 5376,
          "TotalDisk": 9948,
          "AllocatedCpus": 0,
          "AllocatedCpusPercent": 0,
          "AllocatedMemory": 0,
          "AllocatedMemoryPercent": 0,
          "AllocatedDisk": 0,
          "AllocatedDiskPercent": 0,
          "UsedCpus": 0,
          "UsedCpusPercent": 0,
          "UsedMemory": 0,
          "UsedMemoryPercent": 0,
          "UsedDisk": 0,
          "UsedDiskPercent": 0,
          "SlackCpus": 0,
          "SlackCpusPercent": 0,
          "SlackMemory": 0,
          "SlackMemoryPercent": 0,
          "SlackDisk": 0,
          "SlackDiskPercent": 0,
          "Timestamp": 1408319101
        },
        {
          "TotalCpus": 2,
          "TotalMemory": 5000,
          "TotalDisk": 1948,
          "AllocatedCpus": 0,
          "AllocatedCpusPercent": 0,
          "AllocatedMemory": 0,
          "AllocatedMemoryPercent": 0,
          "AllocatedDisk": 0,
          "AllocatedDiskPercent": 0,
          "UsedCpus": 0,
          "UsedCpusPercent": 0,
          "UsedMemory": 0,
          "UsedMemoryPercent": 0,
          "UsedDisk": 0,
          "UsedDiskPercent": 0,
          "SlackCpus": 0,
          "SlackCpusPercent": 0,
          "SlackMemory": 0,
          "SlackMemoryPercent": 0,
          "SlackDisk": 0,
          "SlackDiskPercent": 0,
          "Timestamp": 1408319101
        },
        {
          "TotalCpus": 3,
          "TotalMemory": 376,
          "TotalDisk": 9348,
          "AllocatedCpus": 0,
          "AllocatedCpusPercent": 0,
          "AllocatedMemory": 0,
          "AllocatedMemoryPercent": 0,
          "AllocatedDisk": 0,
          "AllocatedDiskPercent": 0,
          "UsedCpus": 0,
          "UsedCpusPercent": 0,
          "UsedMemory": 0,
          "UsedMemoryPercent": 0,
          "UsedDisk": 0,
          "UsedDiskPercent": 0,
          "SlackCpus": 0,
          "SlackCpusPercent": 0,
          "SlackMemory": 0,
          "SlackMemoryPercent": 0,
          "SlackDisk": 0,
          "SlackDiskPercent": 0,
          "Timestamp": 1408319101
        },
        {
          "TotalCpus": 4,
          "TotalMemory": 5376,
          "TotalDisk": 5548,
          "AllocatedCpus": 0,
          "AllocatedCpusPercent": 0,
          "AllocatedMemory": 0,
          "AllocatedMemoryPercent": 0,
          "AllocatedDisk": 0,
          "AllocatedDiskPercent": 0,
          "UsedCpus": 0,
          "UsedCpusPercent": 0,
          "UsedMemory": 0,
          "UsedMemoryPercent": 0,
          "UsedDisk": 0,
          "UsedDiskPercent": 0,
          "SlackCpus": 0,
          "SlackCpusPercent": 0,
          "SlackMemory": 0,
          "SlackMemoryPercent": 0,
          "SlackDisk": 0,
          "SlackDiskPercent": 0,
          "Timestamp": 1408319101
        }
      ];
      var d3Data = [];
      _.each(STATS, function (key) {
          var series = _.map(response, function (snapshot, i) {
            return {
              x: i,
              y: snapshot[key]
            };
          });
          d3Data.push({name: key, series: series});
      });
      model.set(d3Data);
    // return Backbone.Collection.prototype.fetch.call(this, model);
  },

  url: "/resources"

});
