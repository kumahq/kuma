import{d as y,r as B,c as T,j as n,b as f,F as k,z as m,f as E,h as p,i as c,g as l,E as q,q as S,w as i,e as d,a4 as w,o as t,p as V,m as $}from"./index-0fbacd76.js";import{z as x,c as C,b as L}from"./kongponents.es-c741eab8.js";import{E as N}from"./ErrorBlock-09ceca3c.js";import{_ as z}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-3d8f32b0.js";import{d as I}from"./datadogLogEvents-302eea7b.js";import{Q as _}from"./QueryParameter-70743f73.js";import{_ as W}from"./_plugin-vue_export-helper-c27b6911.js";const O=a=>(V("data-v-a4e17a4a"),a=a(),$(),a),A={class:"tab-container","data-testid":"tab-container"},F={key:0,class:"tab__header"},H={class:"tab__content-container"},Q={class:"flex items-center with-warnings"},j=O(()=>c("span",null,"Warnings",-1)),G=y({__name:"TabsWidget",props:{tabs:{type:Array,required:!0},isLoading:{type:Boolean,required:!1,default:!1},isEmpty:{type:Boolean,required:!1,default:!1},hasError:{type:Boolean,required:!1,default:!1},error:{type:[Error,null],required:!1,default:null},hasBorder:{type:Boolean,required:!1,default:!1},initialTabOverride:{type:String,required:!1,default:null}},emits:["on-tab-change"],setup(a,{emit:b}){const o=a,r=B(""),g=T(()=>o.tabs.map(e=>e.hash.replace("#","")));function h(){const e=_.get("tab");e!==null?r.value=`#${e}`:o.initialTabOverride!==null&&(r.value=`#${o.initialTabOverride}`)}h();function v(e){_.set("tab",e.substring(1)),w.logger.info(I.TABS_TAB_CHANGE,{data:{newActiveTabHash:e}}),b("on-tab-change",e)}return(e,u)=>(t(),n("div",A,[a.isLoading?(t(),f(z,{key:0})):a.error!==null?(t(),f(N,{key:1,error:a.error},null,8,["error"])):(t(),n(k,{key:2},[e.$slots.tabHeader?(t(),n("header",F,[m(e.$slots,"tabHeader",{},void 0,!0)])):E("",!0),p(),c("div",H,[l(d(L),{modelValue:r.value,"onUpdate:modelValue":u[0]||(u[0]=s=>r.value=s),tabs:a.tabs,onChanged:v},q({"warnings-anchor":i(()=>[c("span",Q,[l(d(C),{class:"mr-1",icon:"warning",color:"var(--black-500)","secondary-color":"var(--yellow-300)",size:"16"}),p(),j])]),_:2},[S(g.value,(s,P)=>({name:s,fn:i(()=>[l(d(x),{"border-variant":"noBorder"},{body:i(()=>[m(e.$slots,s,{},void 0,!0)]),_:2},1024)])}))]),1032,["modelValue","tabs"])])],64))]))}});const Y=W(G,[["__scopeId","data-v-a4e17a4a"]]);export{Y as T};
