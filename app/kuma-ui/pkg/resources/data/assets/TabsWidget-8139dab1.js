import{d as v}from"./production-8efaeab1.js";import{m as B,Q as T}from"./kongponents.es-7ead79da.js";import{d as k}from"./datadogLogEvents-302eea7b.js";import{Q as c}from"./QueryParameter-70743f73.js";import{E}from"./ErrorBlock-04954e92.js";import{_ as S}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-ef489e58.js";import{d as q,r as w,c as V,h as n,a as u,F as N,N as f,b as $,f as m,g as i,e as p,a4 as C,m as L,u as l,w as b,o as t,p as x,j as I}from"./runtime-dom.esm-bundler-fd3ecc5a.js";import{_ as Q}from"./_plugin-vue_export-helper-c27b6911.js";const W=a=>(x("data-v-dcbfeb8e"),a=a(),I(),a),O={class:"tab-container","data-testid":"tab-container"},A={key:0,class:"tab__header"},F={class:"tab__content-container"},H={class:"flex items-center with-warnings"},j=W(()=>i("span",null,"Warnings",-1)),z=q({__name:"TabsWidget",props:{tabs:{type:Array,required:!0},isLoading:{type:Boolean,required:!1,default:!1},isEmpty:{type:Boolean,required:!1,default:!1},hasError:{type:Boolean,required:!1,default:!1},error:{type:[Error,null],required:!1,default:null},hasBorder:{type:Boolean,required:!1,default:!1},initialTabOverride:{type:String,required:!1,default:null}},emits:["on-tab-change"],setup(a,{emit:_}){const o=a,r=w(""),g=V(()=>o.tabs.map(e=>e.hash.replace("#","")));function h(){const e=c.get("tab");e!==null?r.value=`#${e}`:o.initialTabOverride!==null&&(r.value=`#${o.initialTabOverride}`)}h();function y(e){c.set("tab",e.substring(1)),v.logger.info(k.TABS_TAB_CHANGE,{data:{newActiveTabHash:e}}),_("on-tab-change",e)}return(e,d)=>(t(),n("div",O,[a.isLoading?(t(),u(S,{key:0})):a.error!==null?(t(),u(E,{key:1,error:a.error},null,8,["error"])):(t(),n(N,{key:2},[e.$slots.tabHeader?(t(),n("header",A,[f(e.$slots,"tabHeader",{},void 0,!0)])):$("",!0),m(),i("div",F,[p(l(T),{modelValue:r.value,"onUpdate:modelValue":d[0]||(d[0]=s=>r.value=s),tabs:a.tabs,onChanged:y},C({"warnings-anchor":b(()=>[i("span",H,[p(l(B),{class:"mr-1",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"16"}),m(),j])]),_:2},[L(l(g),s=>({name:s,fn:b(()=>[f(e.$slots,s,{},void 0,!0)])}))]),1032,["modelValue","tabs"])])],64))]))}});const X=Q(z,[["__scopeId","data-v-dcbfeb8e"]]);export{X as T};
