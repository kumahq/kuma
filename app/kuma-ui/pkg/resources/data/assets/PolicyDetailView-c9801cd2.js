import{_ as n}from"./PolicyDetails.vue_vue_type_script_setup_true_lang-f620cfb4.js";import{e as u,j as _,f as h,_ as y}from"./RouteView.vue_vue_type_script_setup_true_lang-c0a5e54a.js";import{_ as f}from"./RouteTitle.vue_vue_type_script_setup_true_lang-2dc2aa37.js";import{d,c as b,o as s,a as m,w as i,h as p,b as r,g as P,f as x}from"./index-a5906eae.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-4c725a38.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-e1dfd3a9.js";import"./kongponents.es-e59adee0.js";import"./ErrorBlock-c977645b.js";import"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-032c9dac.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-125afdd1.js";import"./TextWithCopyButton-3db8be12.js";import"./TabsWidget-64ce872f.js";import"./QueryParameter-70743f73.js";const z=d({__name:"PolicyDetailView",props:{mesh:{},policyPath:{},policyName:{}},setup(c){const e=c,l=u(),{t:a}=_(),o=b(()=>l.state.policyTypesByPath[e.policyPath]);return(N,w)=>(s(),m(y,null,{default:i(({route:t})=>[p(f,{title:r(a)("policies.routes.item.title",{name:t.params.policy})},null,8,["title"]),P(),p(h,{breadcrumbs:[{to:{name:"policies-list-view",params:{mesh:t.params.mesh,policyPath:t.params.policyPath}},text:r(a)("policies.routes.item.breadcrumbs")}]},{default:i(()=>[o.value?(s(),m(n,{key:0,name:e.policyName,mesh:e.mesh,path:e.policyPath,type:o.value.name},null,8,["name","mesh","path","type"])):x("",!0)]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{z as default};
