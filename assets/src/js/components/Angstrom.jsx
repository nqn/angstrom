/** @jsx React.DOM */

var React = require("react");
var _ = require("underscore");
var Chart = require("./Chart.jsx");
var DataSeries = require("./DataSeries.jsx");
var BackboneMixin = require("../mixins/BackboneMixin");
var d3 = require("d3");
var DataCenterMetricsCollection = require("../models/DataCenterMetricsCollection");

module.exports = React.createClass({
  displayName: "Angstrom",
  mixins: [BackboneMixin],
  getInitialState: function() {
    return {
      collection: new DataCenterMetricsCollection()
    };
  },
  getBackboneModels: function () {
    return this.state.collection;
  },
  getDefaultProps: function () {
    return {
      interpolate: "linear",
      width: 600,
      height: 300
    };
  },
  componentDidMount: function () {
    this.state.collection.fetch({
      reset: true,
      success: function () {
        this._boundForceUpdate();
      }.bind(this)
    });
  },

  render: function() {

    var collection = this.state.collection;
    if (collection.models.length > 0) {
      console.log(collection.get("TotalMemory").get("series"), collection.get("AllocatedMemory").get("series"), collection.get("UsedMemory").get("series"));
      var size = { width: this.props.width, height: this.props.height };

      var max = _.chain(collection.get("TotalMemory").get("series"), collection.get("AllocatedMemory").get("series"), collection.get("UsedMemory").get("series"))
        .zip()
        .map(function(values) {
          return _.reduce(values, function(memo, value) { return Math.max(memo, value.y); }, 0);
        })
        .max()
        .value();

      var xScale = d3.scale.linear()
        .domain([0, collection.get("TotalMemory").get("series").length])
        .range([0, this.props.width]);

      var yScale = d3.scale.linear()
        .domain([0, max])
        .range([0, this.props.height]);
    }

    /* jshint trailing:false, quotmark:false, newcap:false */
    return (
      <div>
        <nav className="navbar navbar-inverse" role="navigation">
         <div className="container-fluid">
            <img src="img/angstrom-2.png" />
            <a className="navbar-brand" href="/">
              Ångström
            </a>
          </div>
        </nav>
        <div className="container-fluid">
          {
            collection.models.length > 0 ?
              <div>
                <Chart width={this.props.width} height={this.props.height}>
                  <DataSeries data={collection.get("TotalMemory").get("series")} size={size} xScale={xScale} yScale={yScale} ref="series1" color="green" fill="green" />

                  <DataSeries data={collection.get("AllocatedMemory").get("series")} size={size} xScale={xScale} yScale={yScale} ref="series2" color="red" fill="red" />

                  <DataSeries data={collection.get("UsedMemory").get("series")} size={size} xScale={xScale} yScale={yScale} ref="series3" color="cornflowerblue" fill="lightsteelblue" />
                </Chart>
              </div> :
              null
          }
        </div>
      </div>
    );
  }
});
