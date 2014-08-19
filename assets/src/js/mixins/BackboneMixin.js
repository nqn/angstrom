module.exports = {
  componentDidMount: function() {
    this._boundForceUpdate = this.forceUpdate.bind(this, null);
    this.getBackboneModels().forEach(function(model) {
      // There are more events that we can listen on. For most cases, we're fetching
      // pages of data, listening to add events causes superfluous calls to render.
      model.on("all", this._boundForceUpdate, this);
      model.fetch({ reset: true });
    }, this);
  },

  componentWillUnmount: function() {
    this.getBackboneModels().forEach(function(model) {
      model.off("all", this._boundForceUpdate, this);
    }, this);
  }
};
