package main

import (
	"ReNaming/internal/renamer"
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ReNamerApp struct {
	App        fyne.App
	MainWindow fyne.Window
	ReNamer    *renamer.ReNamer

	// UI组件
	FileList   *widget.List
	RuleList   *widget.List
	Files      []string
	PreviewBtn *widget.Button
	RenameBtn  *widget.Button
	StatusBar  *widget.Label

	// 存储结果
	previewResults []renamer.ReNameResult
	renameResults  []renamer.ReNameResult
}

func NewReNamerApp() *ReNamerApp {
	a := app.New()
	w := a.NewWindow("ReNamer Lite (仅用于非商业使用)")
	w.Resize(fyne.NewSize(800, 600))

	return &ReNamerApp{
		App:        a,
		MainWindow: w,
		ReNamer:    renamer.NewReNamer(),
		Files:      []string{},
		StatusBar:  widget.NewLabel("0 个文件"),
	}
}

func (r *ReNamerApp) setupUI() {
	// 创建菜单
	r.createMenuBar()

	// 主布局
	mainContainer := container.NewBorder(
		r.createToolbar(),   // 顶部工具栏
		r.createStatusBar(), // 底部状态栏
		nil,
		nil,
		r.createMainContent(), // 中间内容区
	)

	r.MainWindow.SetContent(mainContainer)
}

func (r *ReNamerApp) createMenuBar() {
	// 创建主菜单
	menu := fyne.NewMainMenu(
		fyne.NewMenu("文件",
			fyne.NewMenuItem("添加文件...", r.addFiles),
			fyne.NewMenuItem("添加文件夹...", r.addFolder),
			fyne.NewMenuItem("退出", func() { r.App.Quit() }),
		),
		fyne.NewMenu("设置",
			fyne.NewMenuItem("预览模式", r.toggleDryRun),
		),
		fyne.NewMenu("帮助",
			fyne.NewMenuItem("关于", r.showAbout),
		),
	)

	r.MainWindow.SetMainMenu(menu)
}

func (r *ReNamerApp) createToolbar() fyne.CanvasObject {
	// 工具栏按钮
	addFileBtn := widget.NewButtonWithIcon("添加文件", theme.FileIcon(), r.addFiles)
	addFolderBtn := widget.NewButtonWithIcon("添加文件夹", theme.FolderIcon(), r.addFolder)
	previewBtn := widget.NewButtonWithIcon("预览", theme.SearchIcon(), r.previewRename)
	r.PreviewBtn = previewBtn

	renameBtn := widget.NewButtonWithIcon("重命名", theme.ConfirmIcon(), r.executeRename)
	r.RenameBtn = renameBtn

	// 创建工具栏
	toolbar := container.NewHBox(
		addFileBtn,
		addFolderBtn,
		previewBtn,
		renameBtn,
	)

	return container.NewPadded(toolbar)
}

func (r *ReNamerApp) createMainContent() fyne.CanvasObject {
	// 规则列表
	r.RuleList = widget.NewList(
		func() int { return len(r.ReNamer.Rules) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("规则"),
				widget.NewLabel("说明"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			items := obj.(*fyne.Container).Objects
			rule := r.ReNamer.Rules[id]
			items[0].(*widget.Label).SetText(rule.Name)
			items[1].(*widget.Label).SetText(fmt.Sprintf("%s -> %s", rule.Pattern, rule.Replace))
		},
	)

	// 添加规则按钮
	addRuleBtn := widget.NewButton("点击此处来添加规则", r.showAddRuleDialog)
	ruleContainer := container.NewBorder(nil, addRuleBtn, nil, nil, r.RuleList)

	// 文件列表
	r.FileList = widget.NewList(
		func() int { return len(r.Files) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("状态"),
				widget.NewLabel("旧名"),
				widget.NewLabel("新名"),
				widget.NewLabel("错误信息"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			items := obj.(*fyne.Container).Objects
			filePath := r.Files[id]
			_, fileName := filepath.Split(filePath)

			items[0].(*widget.Label).SetText("")
			items[1].(*widget.Label).SetText(fileName)
			items[2].(*widget.Label).SetText("")
			items[3].(*widget.Label).SetText("")
		},
	)

	// 拖放区域
	dropArea := widget.NewLabel("拖拽文件到此处")
	dropArea.Alignment = fyne.TextAlignCenter

	fileContainer := container.NewBorder(nil, dropArea, nil, nil, r.FileList)

	// 分割视图
	split := container.NewVSplit(
		ruleContainer,
		fileContainer,
	)
	split.Offset = 0.3 // 规则列表占30%高度

	return split
}

func (r *ReNamerApp) createStatusBar() fyne.CanvasObject {
	return container.NewHBox(
		r.StatusBar,
	)
}

// 事件处理函数
func (r *ReNamerApp) addFiles() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}

		filePath := reader.URI().Path()
		// 在Windows上，路径可能以/开头，需要处理
		if strings.HasPrefix(filePath, "/") && len(filePath) > 3 && filePath[2] == ':' {
			filePath = filePath[1:]
		}

		r.Files = append(r.Files, filePath)
		r.ReNamer.AddFiles([]string{filePath})
		r.updateStatusBar()
		r.FileList.Refresh()
	}, r.MainWindow)
}

func (r *ReNamerApp) addFolder() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}

		folderPath := uri.Path()
		// 在Windows上，路径可能以/开头，需要处理
		if strings.HasPrefix(folderPath, "/") && len(folderPath) > 3 && folderPath[2] == ':' {
			folderPath = folderPath[1:]
		}

		// 这里应该遍历文件夹，添加所有文件
		// 简化版本，实际应用中需要更复杂的处理
		dialog.ShowInformation("添加文件夹", "已添加文件夹: "+folderPath, r.MainWindow)
	}, r.MainWindow)
}

func (r *ReNamerApp) showAddRuleDialog() {
	// 创建规则类型选择
	ruleTypes := []string{"添加前缀", "添加后缀", "替换文本", "删除文本", "正则替换"}
	ruleTypeSelect := widget.NewSelect(ruleTypes, nil)

	// 创建输入字段
	patternEntry := widget.NewEntry()
	replaceEntry := widget.NewEntry()

	// 创建表单
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "规则类型", Widget: ruleTypeSelect},
			{Text: "匹配模式", Widget: patternEntry},
			{Text: "替换内容", Widget: replaceEntry},
		},
		OnSubmit: func() {
			// 根据选择的规则类型创建规则
			var rule renamer.Rule
			ruleFactory := renamer.NewRuleFactory()

			switch ruleTypeSelect.Selected {
			case "添加前缀":
				rule = ruleFactory.AddPrefix(replaceEntry.Text)
			case "添加后缀":
				rule = ruleFactory.AddSuffix(replaceEntry.Text)
			case "替换文本":
				rule = ruleFactory.ReplacePattern(patternEntry.Text, replaceEntry.Text)
			case "删除文本":
				rule = ruleFactory.RemovePattern(patternEntry.Text)
			case "正则替换":
				rule = renamer.Rule{
					Name:    "正则替换",
					Pattern: patternEntry.Text,
					Replace: replaceEntry.Text,
				}
			}

			r.ReNamer.AddRule(rule)
			r.RuleList.Refresh()
		},
	}

	dialog.ShowCustom("添加规则", "确定", form, r.MainWindow)
}

func (r *ReNamerApp) previewRename() {
	if len(r.Files) == 0 || len(r.ReNamer.Rules) == 0 {
		dialog.ShowInformation("提示", "请先添加文件和规则", r.MainWindow)
		return
	}

	// 设置为预览模式
	r.ReNamer.SetDryRun(true)

	// 应用规则
	results := r.ReNamer.ApplyBatch()

	// 更新UI显示预览结果
	r.updatePreviewResults(results)
}

func (r *ReNamerApp) executeRename() {
	if len(r.Files) == 0 || len(r.ReNamer.Rules) == 0 {
		dialog.ShowInformation("提示", "请先添加文件和规则", r.MainWindow)
		return
	}

	// 设置为实际执行模式
	r.ReNamer.SetDryRun(false)

	// 应用规则
	results := r.ReNamer.ApplyBatch()

	// 更新UI显示结果
	r.updateRenameResults(results)
}

func (r *ReNamerApp) updatePreviewResults(results []renamer.ReNameResult) {
	// 更新文件列表显示预览结果
	// 不要使用 UpdateAll，而是更新数据源并刷新
	r.Files = make([]string, len(results))
	for i, result := range results {
		r.Files[i] = result.OldPath
	}

	// 保存结果以便在渲染时使用
	r.previewResults = results

	// 刷新列表
	r.FileList.Refresh()
}

func (r *ReNamerApp) updateRenameResults(results []renamer.ReNameResult) {
	// 更新文件列表显示重命名结果
	r.Files = make([]string, 0)
	for _, result := range results {
		if result.Status == renamer.StatusSuccess {
			r.Files = append(r.Files, result.NewPath)
		} else {
			r.Files = append(r.Files, result.OldPath)
		}
	}

	// 保存结果以便在渲染时使用
	r.renameResults = results

	// 刷新列表
	r.FileList.Refresh()

	// 更新状态栏
	r.updateStatusBar()
}

func (r *ReNamerApp) toggleDryRun() {
	// 切换预览模式
	currentMode := r.ReNamer.DryRun
	r.ReNamer.SetDryRun(!currentMode)

	if r.ReNamer.DryRun {
		dialog.ShowInformation("模式切换", "已切换到预览模式，不会实际重命名文件", r.MainWindow)
	} else {
		dialog.ShowInformation("模式切换", "已切换到实际执行模式", r.MainWindow)
	}
}

func (r *ReNamerApp) showAbout() {
	dialog.ShowInformation("关于", "ReNamer Lite v1.0\n仅用于非商业使用\n基于Go语言和Fyne框架开发", r.MainWindow)
}

func (r *ReNamerApp) updateStatusBar() {
	r.StatusBar.SetText(fmt.Sprintf("%d 个文件", len(r.Files)))
}

func main() {
	app := NewReNamerApp()
	app.setupUI()
	app.MainWindow.ShowAndRun()
}
