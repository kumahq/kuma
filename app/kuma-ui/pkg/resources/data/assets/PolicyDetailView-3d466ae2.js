import{_ as n}from"./PolicyDetails.vue_vue_type_script_setup_true_lang-434c43ce.js";import{e as u,j as _,f as h,_ as y}from"./RouteView.vue_vue_type_script_setup_true_lang-ed887f62.js";import{_ as f}from"./RouteTitle.vue_vue_type_script_setup_true_lang-9b835b14.js";import{d,c as b,o as s,a as m,w as i,h as p,b as r,g as P,f as x}from"./index-231ca628.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-2ea01004.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-2ab70067.js";import"./kongponents.es-32169022.js";import"./ErrorBlock-0f3c57b7.js";import"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-fd5e1c64.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-65286b52.js";import"./TextWithCopyButton-289007e3.js";import"./TabsWidget-2ad3bb32.js";import"./QueryParameter-70743f73.js";const z=d({__name:"PolicyDetailView",props:{mesh:{},policyPath:{},policyName:{}},setup(c){const e=c,l=u(),{t:a}=_(),o=b(()=>l.state.policyTypesByPath[e.policyPath]);return(N,w)=>(s(),m(y,null,{default:i(({route:t})=>[p(f,{title:r(a)("policies.routes.item.title",{name:t.params.policy})},null,8,["title"]),P(),p(h,{breadcrumbs:[{to:{name:"policies-list-view",params:{mesh:t.params.mesh,policyPath:t.params.policyPath}},text:r(a)("policies.routes.item.breadcrumbs")}]},{default:i(()=>[o.value?(s(),m(n,{key:0,name:e.policyName,mesh:e.mesh,path:e.policyPath,type:o.value.name},null,8,["name","mesh","path","type"])):x("",!0)]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{z as default};
