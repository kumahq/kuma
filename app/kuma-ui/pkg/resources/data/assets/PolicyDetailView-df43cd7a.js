import{d as N,u as T,r as p,c as y,E as V,j as _,e as s,b as t,w as i,o as e,i as x,t as h,g as v,F as C,q as L,h as k,f as $}from"./index-fb364086.js";import{_ as A}from"./PolicyConnections.vue_vue_type_script_setup_true_lang-f36c8491.js";import{D as S,a as F,_ as j}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-f933dcb8.js";import{E as q}from"./ErrorBlock-5592b912.js";import{_ as H}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-2a0f2006.js";import{T as I}from"./TabsWidget-c7b6fad7.js";import{_ as K}from"./YamlView.vue_vue_type_script_setup_true_lang-d1c12f2f.js";import{u as O}from"./store-ba746ae3.js";import{u as R}from"./index-85d8f0c6.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-90e4bd62.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./kongponents.es-f4d09520.js";import"./datadogLogEvents-302eea7b.js";import"./QueryParameter-70743f73.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-1740043b.js";import"./toYaml-4e00099e.js";const W={class:"policy-details kcard-border"},z={class:"entity-heading","data-testid":"policy-single-entity"},ne=N({__name:"PolicyDetailView",props:{mesh:null,policyPath:null,policyName:null},setup(g){const f=g,w=R(),P=T(),b=O(),D=[{hash:"#overview",title:"Overview"},{hash:"#affected-dpps",title:"Affected DPPs"}],a=p(null),m=p(!0),n=p(null),l=y(()=>{if(a.value===null)return null;const{type:c,name:u,mesh:r}=a.value;return{type:c,name:u,mesh:r}}),d=y(()=>a.value!==null?V(a.value):null);E();function E(){b.dispatch("updatePageTitle",P.params.policy),B(f)}async function B({mesh:c,policyPath:u,policyName:r}){m.value=!0,n.value=null,a.value=null;try{a.value=await w.getSinglePolicyEntity({mesh:c,path:u,name:r})}catch(o){o instanceof Error?n.value=o:console.error(o)}finally{m.value=!1}}return(c,u)=>(e(),_("div",W,[m.value?(e(),s(H,{key:0})):n.value!==null?(e(),s(q,{key:1,error:n.value},null,8,["error"])):t(l)===null?(e(),s(j,{key:2})):(e(),s(I,{key:3,tabs:D},{tabHeader:i(()=>[x("h1",z,h(t(l).name),1)]),overview:i(()=>[v(F,null,{default:i(()=>[(e(!0),_(C,null,L(t(l),(r,o)=>(e(),s(S,{key:o,term:o},{default:i(()=>[k(h(r),1)]),_:2},1032,["term"]))),128))]),_:1}),k(),t(d)!==null?(e(),s(K,{key:0,id:"code-block-policy",class:"mt-4",content:t(d),"is-searchable":""},null,8,["content"])):$("",!0)]),"affected-dpps":i(()=>[v(A,{mesh:t(l).mesh,"policy-name":t(l).name,"policy-type":f.policyPath},null,8,["mesh","policy-name","policy-type"])]),_:1}))]))}});export{ne as default};
