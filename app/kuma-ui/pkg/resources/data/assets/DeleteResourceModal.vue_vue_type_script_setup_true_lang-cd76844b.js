import{d as g,a2 as k,v as B,o as t,g as m,w as p,R as F,l as i,i as n,ab as q,aF as S,j as o,F as y,m as u,D as s,G as T,k as v,aJ as V}from"./index-ba2f01fe.js";const h=["data-testid"],C={key:1},w=g({__name:"DeleteResourceModal",props:{actionButtonText:{type:String,required:!1,default:"Yes, delete"},confirmationText:{type:String,required:!1,default:""},deleteFunction:{type:Function,required:!0},isVisible:{type:Boolean,required:!0},title:{type:String,required:!1,default:"Delete"}},emits:["cancel","delete"],setup(_,{emit:c}){const a=_,{t:d}=k(),e=B(null);async function b(){e.value=null;try{await a.deleteFunction(),c("delete")}catch(r){r instanceof Error?e.value=r:console.error(r)}}return(r,f)=>(t(),m(n(V),{"action-button-text":a.actionButtonText,"confirmation-text":a.confirmationText,"is-visible":a.isVisible,title:a.title,type:"danger",onCanceled:f[0]||(f[0]=l=>c("cancel")),onProceed:b},{"body-content":p(()=>[F(r.$slots,"body-content"),i(),e.value!==null?(t(),m(n(q),{key:0,class:"mt-4",appearance:"danger","is-dismissible":""},{alertMessage:p(()=>[e.value instanceof n(S)?(t(),o(y,{key:0},[u("p",null,s(n(d)("common.error_state.api_error",{status:e.value.status,title:e.value.title})),1),i(),e.value.invalidParameters.length>0?(t(),o("ul",{key:0,"data-testid":`error-${e.value.status}`},[(t(!0),o(y,null,T(e.value.invalidParameters,(l,x)=>(t(),o("li",{key:x},[u("b",null,[u("code",null,s(l.field),1)]),i(": "+s(l.reason),1)]))),128))],8,h)):v("",!0)],64)):(t(),o("p",C,s(n(d)("common.error_state.default_error")),1))]),_:1})):v("",!0)]),_:3},8,["action-button-text","confirmation-text","is-visible","title"]))}});export{w as _};
