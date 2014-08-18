/** @jsx React.DOM */

var React = require("react");

module.exports = React.createClass({
  render: function() {
    return (
      <svg width={this.props.width} height={this.props.height}>{this.props.children}</svg>
    );
  }
});
