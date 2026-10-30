/*
 * End-to-end GUI test for the Playwright MCP tool `browser_run_code_unsafe` (not a @playwright/test spec).
 * Run from Claude Code with the Playwright plugin:
 *   1. cd web && npm run build            (the Go binary embeds web/build)
 *   2. go run .                           (local mode: mock DB + mock KMS on :3000)
 *   3. browser_run_code_unsafe with filename "web/tests/e2e.js"
 * Returns a PASS/FAIL report. Each read opens a fresh page, because navigating to a URL
 * that differs only by its #hash does not reload the app.
 */
async (page) => {
  const BASE = 'http://localhost:3000/';
  const ctx = page.context();
  const results = [];
  const consoleErrors = [];
  const ok = (name, pass, info = '') => results.push(`${pass ? 'PASS' : 'FAIL'} | ${name}${info ? ' | ' + info : ''}`);
  const v = await (await ctx.request.get(BASE + 'variables', { headers: { 'Accept-Language': 'pl' } })).json();

  async function fresh() {
    const p = await ctx.newPage();
    p.alerts = [];
    p.on('dialog', async d => { p.alerts.push(d.message()); await d.accept(); });
    p.on('console', m => { if (m.type() === 'error') consoleErrors.push(m.text().split('\n')[0]); });
    p.on('pageerror', e => consoleErrors.push('PAGEERROR: ' + e.message));
    return p;
  }
  const localize = link => link.replace(/^https?:\/\/[^/]+\//, BASE);
  async function encrypt(msg, password) {
    const p = await fresh();
    await p.goto(BASE);
    await p.locator('textarea').first().fill(msg);
    if (password) { await p.locator('button.btn-success').click(); await p.locator('input[placeholder]').fill(password); }
    await p.locator('button.btn-danger').click();
    await p.locator('#link').waitFor({ timeout: 30000 });
    return { p, link: localize(await p.locator('#link').inputValue()) };
  }
  async function openLink(link) {
    const p = await fresh();
    await p.goto(link);
    await p.locator('button.btn-danger').waitFor();
    return p;
  }
  const shown = async p => { await p.locator('.br-line').waitFor({ timeout: 30000 }); return p.locator('.br-line').textContent(); };
  const readInfo = async p => { await p.getByText(v.messageRead).waitFor({ timeout: 10000 }).catch(() => {}); return p.getByText(v.messageRead).isVisible(); };
  const newBtn = p => p.getByRole('button', { name: v.newMessageButton });

  // 1. empty message
  { const p = await fresh(); await p.goto(BASE);
    await p.locator('button.btn-danger').click(); await p.waitForTimeout(400);
    ok('Pusta wiadomość -> walidacja', await p.locator('textarea.is-invalid').count() === 1 && await p.locator('#link').count() === 0);
    await p.close(); }

  // 2-3. no password round trip
  const msg1 = 'Test bez hasła — zażółć gęślą jaźń 🔐\nlinia 2';
  const { p: e1, link: link1 } = await encrypt(msg1);
  ok('Szyfrowanie bez hasła -> link', link1.includes('#'));
  ok('Przycisk kopiowania obecny', await e1.locator('button.btn-warning').count() === 1);
  await newBtn(e1).click();
  ok('"Nowa wiadomość" z ekranu linku -> formularz', await e1.locator('#link').count() === 0 && await e1.locator('textarea').count() === 1);
  await e1.close();
  { const p = await openLink(link1);
    ok('Link bez hasła: brak pola hasła', await p.locator('input[placeholder]').count() === 0);
    await p.locator('button.btn-danger').click();
    const got = await shown(p);
    ok('Odczyt bez hasła: treść zgodna', got === msg1);
    await newBtn(p).click();
    ok('"Nowa wiadomość" z ekranu treści -> formularz', await p.locator('textarea').count() === 1);
    await p.close(); }

  // 4. second read -> messageRead + button
  { const p = await openLink(link1);
    await p.locator('button.btn-danger').click();
    const info = await readInfo(p);
    ok('Ponowny odczyt -> napis messageRead, bez alertu', info && p.alerts.length === 0, JSON.stringify(p.alerts));
    ok('Ponowny odczyt -> przycisk odczytu ukryty', await p.locator('button.btn-danger').count() === 0);
    await newBtn(p).click();
    ok('"Nowa wiadomość" z ekranu messageRead -> formularz', await p.locator('textarea').count() === 1);
    await p.locator('textarea').fill('z messageRead');
    await p.locator('button.btn-danger').click();
    await p.locator('#link').waitFor({ timeout: 30000 });
    const l = localize(await p.locator('#link').inputValue());
    await p.close();
    const r = await openLink(l); await r.locator('button.btn-danger').click();
    ok('Nowa wiadomość po messageRead da się odczytać', await shown(r) === 'z messageRead');
    await r.close(); }

  // 5. password UI
  { const p = await fresh(); await p.goto(BASE);
    await p.locator('textarea').first().fill('x');
    await p.locator('button.btn-success').click();
    await p.locator('button.btn-danger').click(); await p.waitForTimeout(400);
    ok('Hasło włączone, puste -> walidacja', await p.locator('input[placeholder].is-invalid').count() === 1);
    const color = () => p.locator('svg circle').nth(1).getAttribute('stroke');
    await p.locator('input[placeholder]').fill('a'); await p.waitForTimeout(300); const weak = await color();
    await p.locator('input[placeholder]').fill('Tr0ub4dor&3-horse-battery-staple!'); await p.waitForTimeout(300); const strong = await color();
    ok('Wskaźnik siły hasła', weak !== strong, `${weak} -> ${strong}`);
    await p.locator('input[placeholder] + button').click();
    ok('Pokaż/ukryj hasło', await p.locator('input[placeholder]').getAttribute('type') === 'text');
    await p.locator('button.btn-success').click();
    ok('Ponowny klik "Hasło" chowa pole', !(await p.locator('input[placeholder]').isVisible()));
    await p.close(); }

  // 6. password round trip
  const msg2 = 'Tajne z hasłem 123', pass = 'Kor3ktne-Haslo!';
  const { p: e2, link: link2 } = await encrypt(msg2, pass); await e2.close();
  { const p = await openLink(link2);
    ok('Link z hasłem: pole hasła widoczne', await p.locator('input[placeholder]').count() === 1);
    await p.locator('button.btn-danger').click(); await p.waitForTimeout(1500);
    ok('Puste hasło przy odczycie -> walidacja', await p.locator('input.is-invalid').count() === 1);
    await p.locator('input[placeholder]').fill('zle-haslo');
    await p.locator('button.btn-danger').click();
    await p.locator('button.btn-danger:not(.disabled)').waitFor({ timeout: 30000 }); await p.waitForTimeout(300);
    ok('Błędne hasło -> walidacja, brak treści', await p.locator('input.is-invalid').count() === 1 && await p.locator('.br-line').count() === 0);
    await p.locator('input[placeholder]').fill(pass);
    await p.locator('button.btn-danger').click();
    ok('Poprawne hasło -> treść zgodna', await shown(p) === msg2);
    await p.waitForTimeout(1500); await p.close(); }
  { const p = await openLink(link2);
    await p.locator('input[placeholder]').fill(pass);
    await p.locator('button.btn-danger').click();
    ok('Wiadomość z hasłem usunięta -> messageRead', await readInfo(p) && p.alerts.length === 0 && await p.locator('input[placeholder]').count() === 0);
    await p.close(); }

  // 7. read limit
  { const { p: e, link } = await encrypt('limit', 'Limit-Haslo-99'); await e.close();
    let loads = 0;
    for (let i = 1; i <= 11; i++) {
      const p = await openLink(link);
      await p.locator('input[placeholder]').fill('zle');
      await p.locator('button.btn-danger').click();
      await Promise.race([
        p.locator('input.is-invalid').waitFor({ timeout: 30000 }),
        p.getByText(v.messageRead).waitFor({ timeout: 30000 }),
      ]);
      if (await p.locator('input.is-invalid').count() === 1) loads++;
      await p.close();
    }
    const p = await openLink(link);
    await p.locator('input[placeholder]').fill('Limit-Haslo-99');
    await p.locator('button.btn-danger').click();
    ok('Limit 10 odczytów z hasłem -> potem messageRead', loads === 10 && await readInfo(p), `udane pobrania: ${loads}`);
    await p.close(); }

  // 8. corrupted link
  { const p = await fresh(); await p.goto(BASE + '#@@@nie-base64@@@');
    await p.locator('button.btn-danger').click(); await p.waitForTimeout(800);
    ok('Uszkodzony link -> alert', p.alerts.includes(v.linkIsCorrupted), JSON.stringify(p.alerts));
    await p.close(); }

  // 9. network errors
  { const p = await fresh(); await p.route('**/save', r => r.abort()); await p.goto(BASE);
    await p.locator('textarea').first().fill('net'); await p.locator('button.btn-danger').click(); await p.waitForTimeout(1500);
    ok('Błąd sieci przy zapisie -> alert, przycisk aktywny', p.alerts.includes(v.encryptNetworkError) && await p.locator('button.btn-danger.disabled').count() === 0);
    await p.close(); }
  { const { p: e, link } = await encrypt('read net err'); await e.close();
    const p = await fresh(); await p.route('**/read', r => r.abort()); await p.goto(link);
    await p.locator('button.btn-danger').click(); await p.waitForTimeout(1500);
    ok('Błąd sieci przy odczycie -> alert, bez messageRead', p.alerts.includes(v.decryptNetworkError) && !(await p.getByText(v.messageRead).isVisible()));
    await p.close();
    const p2 = await openLink(link); await p2.locator('button.btn-danger').click();
    ok('Po błędzie sieci wiadomość nadal czytelna', await shown(p2) === 'read net err');
    await p2.close(); }
  { const p = await fresh(); await p.route('**/variables', r => r.abort()); await p.goto(BASE); await p.waitForTimeout(1500);
    ok('Błąd /variables -> alert, strona się renderuje', p.alerts.length === 1 && await p.locator('textarea').count() === 1);
    ok('Błąd /variables -> domyślny tytuł i lang', await p.title() !== '' && await p.locator('html').getAttribute('lang') === 'en');
    await p.close(); }

  // 10. long message
  { const long = 'A'.repeat(100000);
    const { p: e, link } = await encrypt(long); await e.close();
    const p = await openLink(link); await p.locator('button.btn-danger').click();
    ok('Długa wiadomość (100k znaków)', (await shown(p)) === long);
    await p.close(); }

  // 11. English locale
  { const c = await ctx.browser().newContext({ locale: 'en-US' });
    const p = await c.newPage(); await p.goto(BASE);
    const btn = await p.locator('button.btn-danger').innerText();
    ok('Wersja angielska', /[A-Za-z]/.test(btn) && btn !== v.secureButton, btn.trim());
    await c.close(); }

  // 12. page title and lang follow the served language
  for (const [header, lang] of [['en-US', 'en'], ['pl-PL', 'pl'], ['zh-TW', 'zh-Hans'], ['nn-NO', 'nb']]) {
    const c = await ctx.browser().newContext({ extraHTTPHeaders: { 'Accept-Language': header } });
    const expected = await (await c.request.get(BASE + 'variables')).json();
    const p = await c.newPage(); await p.goto(BASE); await p.locator('textarea').first().waitFor();
    const title = await p.title(), htmlLang = await p.locator('html').getAttribute('lang');
    ok(`Tytuł i lang dla ${header}`, title === expected.title && htmlLang === lang, `${htmlLang} | ${title}`);
    await c.close();
  }

  const failed = results.filter(r => r.startsWith('FAIL')).length;
  return `${results.length - failed}/${results.length} PASS\n` + results.join('\n') +
    '\n\nBłędy konsoli:\n' + ([...new Set(consoleErrors)].join('\n') || '(brak)');
}
