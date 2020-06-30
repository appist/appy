const { I } = inject();

let docURL;

Given("I know the documentation URL is {string}", url => {
  docURL = url;
});

When(/I open the URL/, () => {
  I.amOnPage(docURL);
});

Then(/I should see the introduction page/, () => {
  I.see("Introduction");
});
