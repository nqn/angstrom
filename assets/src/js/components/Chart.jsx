/** @jsx React.DOM */

var React = require("react");
var _ = require("underscore");
// var d3 = require("d3");
var Rickshaw = require("rickshaw");
// require("../libs/nv.d3");

module.exports = React.createClass({
  displayName: "Chart",

  propTypes: {
    definitions: React.PropTypes.object.isRequired,
    collection: React.PropTypes.object.isRequired,
    onSelectApp: React.PropTypes.func.isRequired
  },

  getInitialState: function() {
    return {
      chart: null
    };
  },

  getDefaultProps: function () {
    return {
      definitions: [],
      collection: [],
      width: 1200,
      height: 400
    };
  },

  componentDidMount: function() {
    var chart = new Rickshaw.Graph( {
      element: this.getDOMNode(),
      width: this.props.width,
      height: this.props.height,
      renderer: "bar",
      // renderer: "area",
      // orientation: "left",
      offset: "stack",
      stroke: true,
      preserve: true,
      interpolation: "step-after",
      // interpolation: "linear",
      // series: this.props.collection

      // series: new Rickshaw.Series(
      //   this.props.collection,
      //   undefined,
      //   {}
      // )

      series: new Rickshaw.Series.FixedDuration(
        // this.props.collection,
        this.props.definitions,
        undefined,
        {
          timeInterval: 5000,
          maxDataPoints: 100//,
          // timeBase: new Date().getTime() / 1000
        }
      )
    });

    chart.render();




    // var preview = new Rickshaw.Graph.RangeSlider( {
    //   graph: graph,
    //   element: document.getElementById('preview'),
    // } );

    var hoverDetail = new Rickshaw.Graph.HoverDetail( {
      graph: chart,
      xFormatter: function(x) {
        return new Date(x * 1000).toString();
      }
    } );

    // var annotator = new Rickshaw.Graph.Annotate( {
    //   graph: graph,
    //   element: document.getElementById('timeline')
    // } );

    // var legend = new Rickshaw.Graph.Legend( {
    //   graph: graph,
    //   element: document.getElementById('legend')

    // } );

    // var shelving = new Rickshaw.Graph.Behavior.Series.Toggle( {
    //   graph: graph,
    //   legend: legend
    // } );

    // var order = new Rickshaw.Graph.Behavior.Series.Order( {
    //   graph: graph,
    //   legend: legend
    // } );

    // var highlighter = new Rickshaw.Graph.Behavior.Series.Highlight( {
    //   graph: graph,
    //   legend: legend
    // } );

    // var smoother = new Rickshaw.Graph.Smoother( {
    //   graph: graph,
    //   element: document.querySelector('#smoother')
    // } );

    var ticksTreatment = "glow";

    var xAxis = new Rickshaw.Graph.Axis.Time( {
      graph: chart,
      ticksTreatment: ticksTreatment,
      timeFixture: new Rickshaw.Fixtures.Time.Local()
    });

    xAxis.render();

    var yAxis = new Rickshaw.Graph.Axis.Y( {
      graph: chart,
      tickFormat: Rickshaw.Fixtures.Number.formatKMBT,
      ticksTreatment: ticksTreatment
    });

    yAxis.render();


    // var controls = new RenderControls( {
    //   element: document.querySelector('form'),
    //   graph: graph
    // } );

    // set state
    this.setState({chart: chart});
  },

  componentWillUnmount: function() {
    this.setState({chart: null});
  },

  shouldComponentUpdate: function(props) {
    if (this.state.chart != null) {
      _.each(props.collection, function (dataPoint) {
        // get only data in defintion
        var data = {};
        _.each(this.props.definitions, function (def) {
          data[def.name] = dataPoint[def.name];
        });
        // data.timeBase = dataPoint.Timestamp;
        this.state.chart.series.addData(data);
      }, this);
      this.state.chart.render();
    }
    // always skip React's render step
    return false;
  },

  render: function() {
    return (
      <div />
    );
  }
});
