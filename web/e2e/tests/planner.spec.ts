import { test, expect } from '@playwright/test';

/**
 * Eino 自动规划（Planner）页面冒烟：与 DeerFlow 对齐计划中的「前端 / 规划入口」回归锚点。
 */
test.test('Planner 页面', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/planner');
    await page.waitForSelector('.ant-card', { timeout: 10000 });
  });

  test('页面加载成功', async ({ page }) => {
    await expect(page.locator('text=Eino 自动规划系统')).toBeVisible();
  });
});
