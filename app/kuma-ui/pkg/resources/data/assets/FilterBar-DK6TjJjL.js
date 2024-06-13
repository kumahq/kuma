var J=Object.defineProperty;var Q=(e,i,t)=>i in e?J(e,i,{enumerable:!0,configurable:!0,writable:!0,value:t}):e[i]=t;var k=(e,i,t)=>(Q(e,typeof i!="symbol"?i+"":i,t),t);import{d as j,I as B,c as _,o as u,a as W,w as P,b as f,t as y,e as N,n as Y,r as X,p as K,f as O,g as a,s as ee,G as b,a6 as te,aq as ie,ar as re,i as se,j as C,Z as ne,k as h,A as I,as as oe,K as q,h as ae,at as le,au as ue,l as D,H as ce,J as de,_ as fe}from"./index-B1qFrf1M.js";const ge=e=>(K("data-v-678df7f6"),e=e(),O(),e),pe=["aria-hidden"],he={key:0,"data-testid":"kui-icon-svg-title"},me=ge(()=>a("path",{d:"M9.4 18L8 16.6L12.6 12L8 7.4L9.4 6L15.4 12L9.4 18Z",fill:"currentColor"},null,-1)),ve=j({__name:"ChevronRightIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:B,validator:e=>{if(typeof e=="number"&&e>0)return!0;if(typeof e=="string"){const i=String(e).replace(/px/gi,""),t=Number(i);if(t&&!isNaN(t)&&Number.isInteger(t)&&t>0)return!0}return!1}},as:{type:String,required:!1,default:"span"}},setup(e){const i=e,t=_(()=>{if(typeof i.size=="number"&&i.size>0)return`${i.size}px`;if(typeof i.size=="string"){const c=String(i.size).replace(/px/gi,""),o=Number(c);if(o&&!isNaN(o)&&Number.isInteger(o)&&o>0)return`${o}px`}return B}),g=_(()=>({boxSizing:"border-box",color:i.color,display:i.display,height:t.value,lineHeight:"0",width:t.value}));return(c,o)=>(u(),W(X(e.as),{"aria-hidden":e.decorative?"true":void 0,class:"kui-icon chevron-right-icon","data-testid":"kui-icon-wrapper-chevron-right-icon",style:Y(g.value)},{default:P(()=>[(u(),f("svg",{"aria-hidden":e.decorative?"true":void 0,"data-testid":"kui-icon-svg-chevron-right-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg"},[e.title?(u(),f("title",he,y(e.title),1)):N("",!0),me],8,pe))]),_:1},8,["aria-hidden","style"]))}}),be=ee(ve,[["__scopeId","data-v-678df7f6"]]),ye=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class Se{constructor(i,t){k(this,"commands");k(this,"keyMap");k(this,"boundTriggerShortcuts");this.commands=t,this.keyMap=Object.fromEntries(Object.entries(i).map(([g,c])=>[g.toLowerCase(),c])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(i){_e(i,this.keyMap,this.commands)}}function _e(e,i,t){const g=we(e.code),c=[e.ctrlKey?"ctrl":"",e.shiftKey?"shift":"",e.altKey?"alt":"",g].filter(S=>S!=="").join("+"),o=i[c];if(!o)return;const p=t[o];p.isAllowedContext&&!p.isAllowedContext(e)||(p.shouldPreventDefaultAction&&e.preventDefault(),!(p.isDisabled&&p.isDisabled())&&p.trigger(e))}function we(e){return ye.includes(e)?"":e.replace(/^Key/,"").toLowerCase()}let M=0;const ke=(e="unique")=>(M++,`${e}-${M}`),Ie=e=>(K("data-v-7a0290e4"),e=e(),O(),e),xe=Ie(()=>a("span",{class:"visually-hidden"},"Focus filter",-1)),Ce={class:"filter-bar-icon"},Ne=["for"],Le=["id","placeholder"],Fe={key:0,class:"suggestion-box","data-testid":"filter-bar-suggestion-box"},Te={class:"suggestion-list"},$e={key:0,class:"filter-bar-error"},Ae={key:0},ze=["title","data-filter-field"],Be={class:"visually-hidden"},qe=j({__name:"FilterBar",props:{fields:{},placeholder:{default:""},query:{default:""},id:{default:()=>ke("filter-bar")}},emits:["change"],setup(e,{emit:i}){const t=e,g=b(),c=i,o=r=>{r!=null&&r.target&&(c("change",new FormData(r.target)),m.value=!1)},p=r=>{c("change",new FormData(g.value))},S=b(null),d=b(null),L=b(null),m=b(!1),v=b(t.query);te(()=>t.query,r=>{v.value=r});const w=b(0),F=_(()=>Object.keys(t.fields)),T=_(()=>Object.entries(t.fields).slice(0,5).map(([r,s])=>({fieldName:r,...s}))),$=_(()=>F.value.length>0?`Filter by ${F.value.join(", ")}`:"Filter"),E=_(()=>t.placeholder??$.value),R={ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},H={jumpToNextSuggestion:{trigger:()=>z(1),isAllowedContext(r){return d.value!==null&&r.composedPath().includes(d.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:()=>z(-1),isAllowedContext(r){return d.value!==null&&r.composedPath().includes(d.value)},shouldPreventDefaultAction:!0}},A=new Se(R,H);ie(function(){A.registerListener()}),re(function(){A.unRegisterListener()});function z(r){const s=T.value.length;let l=w.value+r;l===-1&&(l=s),w.value=l%(s+1)}function V(){d.value instanceof HTMLInputElement&&d.value.focus()}function U(r){const l=r.currentTarget.getAttribute("data-filter-field");l&&d.value instanceof HTMLInputElement&&Z(d.value,l)}function Z(r,s){const l=v.value===""||v.value.endsWith(" ")?"":" ";v.value+=l+s+":",r.focus(),w.value=0}function G(r){r.relatedTarget===null&&(m.value=!1),S.value instanceof HTMLElement&&r.relatedTarget instanceof Node&&!S.value.contains(r.relatedTarget)&&(m.value=!1)}return(r,s)=>{const l=se("search");return u(),f("div",{ref_key:"filterBar",ref:S,class:"filter-bar","data-testid":"filter-bar"},[C(l,null,{default:P(()=>[a("form",{ref_key:"$form",ref:g,onSubmit:ne(o,["prevent"])},[a("button",{class:"focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"filter-bar-focus-filter-input-button",onClick:V},[xe,h(),a("span",Ce,[C(I(oe),{decorative:"","data-testid":"filter-bar-filter-icon","hide-title":"",size:I(q)},null,8,["size"])])]),h(),a("label",{for:`${t.id}-filter-bar-input`,class:"visually-hidden"},[ae(r.$slots,"default",{},()=>[h(y($.value),1)],!0)],8,Ne),h(),le(a("input",{id:`${t.id}-filter-bar-input`,ref_key:"filterInput",ref:d,"onUpdate:modelValue":s[0]||(s[0]=n=>v.value=n),class:"filter-bar-input",type:"search",placeholder:E.value,"data-testid":"filter-bar-filter-input",name:"s",onFocus:s[1]||(s[1]=n=>m.value=!0),onInput:s[2]||(s[2]=n=>m.value=!0),onBlur:G,onSearch:s[3]||(s[3]=n=>{n.target.value.length===0&&(p(n),m.value=!0)})},null,40,Le),[[ue,v.value]]),h(),m.value?(u(),f("div",Fe,[a("div",Te,[L.value!==null?(u(),f("p",$e,y(L.value.message),1)):(u(),f("button",{key:1,type:"submit",class:D(["submit-query-button",{"submit-query-button-is-selected":w.value===0}]),"data-testid":"filter-bar-submit-query-button"},`
              Submit `+y(v.value),3)),h(),(u(!0),f(ce,null,de(T.value,(n,x)=>(u(),f("div",{key:`${t.id}-${x}`,class:D(["suggestion-list-item",{"suggestion-list-item-is-selected":w.value===x+1}])},[a("b",null,y(n.fieldName),1),n.description!==""?(u(),f("span",Ae,": "+y(n.description),1)):N("",!0),h(),a("button",{class:"apply-suggestion-button",title:`Add ${n.fieldName}:`,type:"button","data-filter-field":n.fieldName,"data-testid":"filter-bar-apply-suggestion-button",onClick:U},[a("span",Be,"Add "+y(n.fieldName)+":",1),h(),C(I(be),{decorative:"","hide-title":"",size:I(q)},null,8,["size"])],8,ze)],2))),128))])])):N("",!0)],544)]),_:3})],512)}}}),je=fe(qe,[["__scopeId","data-v-7a0290e4"]]);export{je as F};
