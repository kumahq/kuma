import{d as k,l as B,D as q,o as t,b as m,w as p,r as h,f as i,p as n,a5 as F,W as S,c as o,F as v,m as u,t as l,G as T,q as y,a6 as V}from"./index-iznqiN1v.js";const C=["data-testid"],D={key:1},w=k({__name:"DeleteResourceModal",props:{actionButtonText:{type:String,required:!1,default:"Yes, delete"},confirmationText:{type:String,required:!1,default:""},deleteFunction:{type:Function,required:!0},isVisible:{type:Boolean,required:!0},title:{type:String,required:!1,default:"Delete"}},emits:["cancel","delete"],setup(_,{emit:x}){const{t:c}=B(),a=_,d=x,e=q(null);async function b(){e.value=null;try{await a.deleteFunction(),d("delete")}catch(r){r instanceof Error?e.value=r:console.error(r)}}return(r,f)=>(t(),m(n(V),{"action-button-text":a.actionButtonText,"confirmation-text":a.confirmationText,visible:a.isVisible,title:a.title,type:"danger",onCancel:f[0]||(f[0]=s=>d("cancel")),onProceed:b},{default:p(()=>[h(r.$slots,"default"),i(),e.value!==null?(t(),m(n(F),{key:0,class:"mt-4",appearance:"danger"},{default:p(()=>[e.value instanceof n(S)?(t(),o(v,{key:0},[u("p",null,l(n(c)("common.error_state.api_error",{status:e.value.status,title:e.value.detail})),1),i(),e.value.invalidParameters.length>0?(t(),o("ul",{key:0,"data-testid":`error-${e.value.status}`},[(t(!0),o(v,null,T(e.value.invalidParameters,(s,g)=>(t(),o("li",{key:g},[u("b",null,[u("code",null,l(s.field),1)]),i(": "+l(s.reason),1)]))),128))],8,C)):y("",!0)],64)):(t(),o("p",D,l(n(c)("common.error_state.default_error")),1))]),_:1})):y("",!0)]),_:3},8,["action-button-text","confirmation-text","visible","title"]))}});export{w as _};
