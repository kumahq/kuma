import{d as v,r as B,i as k,o as t,c as l,j as f,F as T,R as p,y as E,e as _,f as c,a as i,Q as S,A as q,w as d,u as o,I as w,b as V,ak as I,af as N,ag as $,N as x,O as C,J as L}from"./index-5f1fbf13.js";import{E as O}from"./ErrorBlock-7fdfff5d.js";import{_ as W}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-48d7a99c.js";import{Q as b}from"./QueryParameter-70743f73.js";const A=a=>(x("data-v-a4e17a4a"),a=a(),C(),a),Q={class:"tab-container","data-testid":"tab-container"},F={key:0,class:"tab__header"},H={class:"tab__content-container"},j={class:"flex items-center with-warnings"},z=A(()=>c("span",null,"Warnings",-1)),G=v({__name:"TabsWidget",props:{tabs:{type:Array,required:!0},isLoading:{type:Boolean,required:!1,default:!1},isEmpty:{type:Boolean,required:!1,default:!1},hasError:{type:Boolean,required:!1,default:!1},error:{type:[Error,null],required:!1,default:null},hasBorder:{type:Boolean,required:!1,default:!1},initialTabOverride:{type:String,required:!1,default:null}},emits:["on-tab-change"],setup(a,{emit:m}){const n=a,s=B(""),g=k(()=>n.tabs.map(e=>e.hash.replace("#","")));function y(){const e=b.get("tab");e!==null?s.value=`#${e}`:n.initialTabOverride!==null&&(s.value=`#${n.initialTabOverride}`)}y();function h(e){b.set("tab",e.substring(1)),N.logger.info($.TABS_TAB_CHANGE,{data:{newActiveTabHash:e}}),m("on-tab-change",e)}return(e,u)=>(t(),l("div",Q,[a.isLoading?(t(),f(W,{key:0})):a.error!==null?(t(),f(O,{key:1,error:a.error},null,8,["error"])):(t(),l(T,{key:2},[e.$slots.tabHeader?(t(),l("header",F,[p(e.$slots,"tabHeader",{},void 0,!0)])):E("",!0),_(),c("div",H,[i(o(I),{modelValue:s.value,"onUpdate:modelValue":u[0]||(u[0]=r=>s.value=r),tabs:a.tabs,onChanged:h},S({"warnings-anchor":d(()=>[c("span",j,[i(o(V),{class:"mr-1",icon:"warning",color:"var(--black-500)","secondary-color":"var(--yellow-300)",size:"16"}),_(),z])]),_:2},[q(o(g),(r,J)=>({name:r,fn:d(()=>[i(o(w),{"border-variant":"noBorder"},{body:d(()=>[p(e.$slots,r,{},void 0,!0)]),_:2},1024)])}))]),1032,["modelValue","tabs"])])],64))]))}});const K=L(G,[["__scopeId","data-v-a4e17a4a"]]);export{K as T};
