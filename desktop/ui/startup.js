window.envariumStartupError = (detail) => {
  document.querySelector('#status').textContent = 'Envarium could not start safely.';
  const output = document.querySelector('#detail');
  output.textContent = detail;
  output.hidden = false;
};
