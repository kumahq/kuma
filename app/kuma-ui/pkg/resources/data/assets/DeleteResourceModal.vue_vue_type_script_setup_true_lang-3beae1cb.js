import{d as k,l as B,m as q,o as t,b as f,w as p,r as C,f as l,q as n,$ as F,a5 as S,c as s,F as y,p as u,t as o,I as T,s as v,aC as V}from"./index-21b00685.js";const $=["data-testid"],h={key:1},D=k({__name:"DeleteResourceModal",props:{actionButtonText:{type:String,required:!1,default:"Yes, delete"},confirmationText:{type:String,required:!1,default:""},deleteFunction:{type:Function,required:!0},isVisible:{type:Boolean,required:!0},title:{type:String,required:!1,default:"Delete"}},emits:["cancel","delete"],setup(_,{emit:b}){const{t:c}=B(),a=_,d=b,e=q(null);async function x(){e.value=null;try{await a.deleteFunction(),d("delete")}catch(r){r instanceof Error?e.value=r:console.error(r)}}return(r,m)=>(t(),f(n(V),{"action-button-text":a.actionButtonText,"confirmation-text":a.confirmationText,"is-visible":a.isVisible,title:a.title,type:"danger",onCanceled:m[0]||(m[0]=i=>d("cancel")),onProceed:x},{"body-content":p(()=>[C(r.$slots,"body-content"),l(),e.value!==null?(t(),f(n(F),{key:0,class:"mt-4",appearance:"danger","is-dismissible":""},{alertMessage:p(()=>[e.value instanceof n(S)?(t(),s(y,{key:0},[u("p",null,o(n(c)("common.error_state.api_error",{status:e.value.status,title:e.value.detail})),1),l(),e.value.invalidParameters.length>0?(t(),s("ul",{key:0,"data-testid":`error-${e.value.status}`},[(t(!0),s(y,null,T(e.value.invalidParameters,(i,g)=>(t(),s("li",{key:g},[u("b",null,[u("code",null,o(i.field),1)]),l(": "+o(i.reason),1)]))),128))],8,$)):v("",!0)],64)):(t(),s("p",h,o(n(c)("common.error_state.default_error")),1))]),_:1})):v("",!0)]),_:3},8,["action-button-text","confirmation-text","is-visible","title"]))}});export{D as _};
